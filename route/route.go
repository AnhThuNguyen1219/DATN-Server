package route

import (
	"errors"
	"log"
	"net/http"
	"regexp"

	"server/backend/auth"
	"server/backend/database"
	"server/backend/models"
	"server/backend/utils"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v7"
	"github.com/julienschmidt/httprouter"
)

func Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params, client *redis.Client) {
	// Convert json data from request to type User
	var user models.User
	err := utils.DecodeJSONBody(w, r, &user)
	if err != nil {
		var mr *utils.MalformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if govalidator.IsNull(user.Username) || govalidator.IsNull(user.Password) {
		utils.JSON(w, http.StatusBadRequest, "Input cannot be empty")
		return
	}

	// Escape string before query to sql server
	user.Username = models.Santize(user.Username)
	user.Password = models.Santize(user.Password)

	// Check with regex
	match, err := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9]{4,255}", user.Username)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if !match {
		utils.JSON(w, http.StatusBadRequest, "Username must longer than 3 characters can only include: a-z, A-Z, 0-9")
		return
	}

	match, err = regexp.MatchString(".{5,255}$", user.Password)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if !match {
		utils.JSON(w, http.StatusBadRequest, "Password length must between 5 and 255 characters")
		return
	}

	db := database.Connect()
	isUsernameExist := database.IsUserExist(db, user.Username)

	if isUsernameExist == false {
		utils.JSON(w, http.StatusBadRequest, "Username or Password incorrect") // Username is not exist, but show incorrect
		return
	}

	hashedPassword := database.GetHashedPassword(db, user.Username)

	check := models.CheckPasswordHash(hashedPassword, user.Password)

	if check != true {
		utils.JSON(w, http.StatusUnauthorized, "Username or Password incorrect")
		return
	}

	token, errCreate := auth.CreateToken(user.Username)

	if errCreate != nil {
		log.Println(errCreate.Error())
		utils.JSON(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	saveErr := auth.SaveAuthRedis(client, user.Username, token)
	if saveErr != nil {
		utils.JSON(w, http.StatusUnprocessableEntity, saveErr.Error())
		return
	}

	utils.JSON(w, http.StatusOK, struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	})
}

func RefreshTokenAPI(w http.ResponseWriter, r *http.Request, _ httprouter.Params, client *redis.Client) {
	token, status, statusMsg := auth.RefreshToken(client, r)
	if token != nil {
		// Create new token successfully
		utils.JSON(w, status, struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
		}{
			AccessToken:  token["access_token"],
			RefreshToken: token["refresh_token"],
		})
	} else {
		utils.JSON(w, status, struct {
			ErrorMsg string `json:"error"`
		}{
			ErrorMsg: statusMsg,
		})
	}
}

//Content Text API
type CreateContentTextInfor struct {
	Text string `json:"content_text"`
}

func CreateContentTextAPI(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get data from json request
	var data CreateContentTextInfor
	err := utils.DecodeJSONBody(w, r, &data)
	if err != nil {
		var mr *utils.MalformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	db := database.Connect()
	err = database.CreateContentText(db, data.Text, data.Text)
	if err != nil {
		log.Println(err.Error())
		utils.JSON(w, http.StatusInternalServerError, "Can't create new text content")
		return
	}
}

//Content Image API
type CreateContentImageInfor struct {
	Title          string `json: "content_title"'`
	OriginalImgURL string `json:"original_img_url"`
	PreviewImgURL  string `json:"preview_img_url"`
}

func CreateContentImageAPI(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get data from json request
	var data CreateContentImageInfor
	err := utils.DecodeJSONBody(w, r, &data)
	if err != nil {
		var mr *utils.MalformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	db := database.Connect()
	err = database.CreateContentImage(db, data.Title, data.OriginalImgURL, data.PreviewImgURL)
	if err != nil {
		log.Println(err.Error())
		utils.JSON(w, http.StatusInternalServerError, "Can't create new message")
		return
	}
}
