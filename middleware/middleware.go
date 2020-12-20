package middleware

import (
	"net/http"

	"server/backend/utils"

	"github.com/go-redis/redis/v7"
	"github.com/julienschmidt/httprouter"
)

type RedisHandle func(http.ResponseWriter, *http.Request, httprouter.Params, *redis.Client)

func AuthMiddleware(f httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
		// isValid, err := auth.IsTokenValid(r)
		// if err != nil {
		// 	utils.JSON(w, http.StatusInternalServerError, "Internal Server Erroraaa")
		// 	return
		// }
		// if !isValid {
		// 	utils.JSON(w, http.StatusUnauthorized, "Token invalid")
		// 	return
		// }
		if r.Header.Get("Authorization") == "" {
			utils.JSON(w, http.StatusInternalServerError, "Internal Server Erroraaa")
			return
		}
		f(w, r, param)
	}
}
func AddRedisClientMiddleware(f RedisHandle, client *redis.Client) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
		f(w, r, param, client)
	}
}
