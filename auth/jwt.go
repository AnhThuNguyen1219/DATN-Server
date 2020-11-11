package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v7"
	"github.com/twinj/uuid"
)

type TokenDetail struct {
	AccessToken    string
	RefreshToken   string
	AccessUuid     string
	RefreshUuid    string
	AccessExpires  int64
	RefreshExpires int64
}

func CreateToken(username string) (*TokenDetail, error) {
	tokenDetail := &TokenDetail{}
	tokenDetail.AccessExpires = time.Now().Add(time.Minute * 15).Unix()
	tokenDetail.AccessUuid = uuid.NewV4().String()
	tokenDetail.RefreshExpires = time.Now().Add(time.Hour * 24 * 3).Unix()
	tokenDetail.RefreshUuid = uuid.NewV4().String()

	var err error

	// Creating access token
	accessTokenClaims := jwt.MapClaims{}
	accessTokenClaims["authorized"] = true
	accessTokenClaims["username"] = username
	accessTokenClaims["access_uuid"] = tokenDetail.AccessUuid
	accessTokenClaims["exp"] = tokenDetail.AccessExpires
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	tokenDetail.AccessToken, err = accessToken.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	// Creating refresh token
	refreshTokenClaims := jwt.MapClaims{}
	refreshTokenClaims["refresh_uuid"] = tokenDetail.AccessUuid
	refreshTokenClaims["username"] = username
	refreshTokenClaims["exp"] = tokenDetail.RefreshExpires
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	tokenDetail.RefreshToken, err = refreshToken.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return nil, err
	}

	return tokenDetail, nil
}

// SaveAuthRedis save authentication info into Redis, set the expiration time for each record in redis
// in order to delete it automatically
func SaveAuthRedis(client *redis.Client, username string, tokenDetail *TokenDetail) error {
	accessToken := time.Unix(tokenDetail.AccessExpires, 0) //converting Unix to UTC(to Time object)
	refreshToken := time.Unix(tokenDetail.RefreshExpires, 0)
	now := time.Now()

	errAccess := client.Set(tokenDetail.AccessUuid, username, accessToken.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}
	errRefresh := client.Set(tokenDetail.RefreshUuid, username, refreshToken.Sub(now)).Err()
	if errRefresh != nil {
		return errRefresh
	}

	return nil
}

func ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	//normally Authorization the_token_xxx
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func IsTokenValid(r *http.Request) (bool, error) {
	token, err := VerifyToken(r)
	if err != nil {
		return false, err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return false, err
	}
	return true, nil
}

type AccessDetail struct {
	AccessUuid string
	Username   string
}

func ExtractTokenMetadata(r *http.Request) (*AccessDetail, error) {
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		username, ok := claims["username"].(string)
		if !ok {
			return nil, err
		}
		return &AccessDetail{
			AccessUuid: accessUuid,
			Username:   username,
		}, nil
	}
	return nil, err
}

func FetchAuthRedis(client *redis.Client, authDetail *AccessDetail) (string, error) {
	username, err := client.Get(authDetail.AccessUuid).Result()
	if err != nil {
		return "", err
	}
	return username, nil
}

func DeleteAuthRedis(client *redis.Client, givenUuid string) (int64, error) {
	deleted, err := client.Del(givenUuid).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

func RefreshToken(client *redis.Client, r *http.Request) (map[string]string, int, string) {
	mapToken := map[string]string{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		return nil, http.StatusBadRequest, "Can't read request body"
	}
	err = json.Unmarshal(body, &mapToken)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err.Error()
	}
	refreshToken := mapToken["refresh_token"]

	//verify the token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})
	//if there is an error, the token must have expired
	if err != nil {
		return nil, http.StatusUnauthorized, "Refresh token expired"
	}
	//is token valid?
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return nil, http.StatusUnauthorized, "Token invalid or token expired"
	}
	//Since token is valid, get the uuid:
	claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims
	if ok && token.Valid {
		refreshUuid, ok := claims["refresh_uuid"].(string) //convert the interface to string
		if !ok {
			return nil, http.StatusUnprocessableEntity, "Error occurred"
		}
		username, ok := claims["username"].(string)
		if !ok {
			return nil, http.StatusUnprocessableEntity, "Error occurred"
		}
		//Delete the previous Refresh Token
		deleted, delErr := DeleteAuthRedis(client, refreshUuid)
		if delErr != nil || deleted == 0 { //if anything goes wrong
			return nil, http.StatusInternalServerError, "Internal Server Error"
		}
		//Create new pairs of refresh and access tokens
		newPairToken, createErr := CreateToken(username)
		if createErr != nil {
			return nil, http.StatusForbidden, createErr.Error()
		}
		saveErr := SaveAuthRedis(client, username, newPairToken)
		if saveErr != nil {
			return nil, http.StatusForbidden, saveErr.Error()
		}
		tokens := map[string]string{
			"access_token":  newPairToken.AccessToken,
			"refresh_token": newPairToken.RefreshToken,
		}

		return tokens, http.StatusCreated, ""
	} else {
		return nil, http.StatusUnauthorized, "Refresh token expired"
	}
}
