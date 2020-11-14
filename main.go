package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"server/backend/middleware"
	"server/backend/route"

	"github.com/go-redis/redis/v7"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

var (
	client *redis.Client
)

func init() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if len(redisAddr) == 0 {
		redisAddr = "localhost:6379"
	}
	fmt.Println(redisAddr)
	client = redis.NewClient(&redis.Options{
		Addr:     "redis-11192.c91.us-east-1-3.ec2.cloud.redislabs.com:11192", // use default Addr
		Password: "PQeZwidqUYvUNAJ9aVlLvI98fr8u39tf",                          // no password set
		DB:       0,                                                           // use default DB
	})
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}
}

func main() {
	router := httprouter.New()

	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Fprint(w, "<h1>Welcome</h1>")
	})
	router.POST("/auth/login", middleware.AddRedisClientMiddleware(route.Login, client))
	router.POST("/api/token/refresh", middleware.AddRedisClientMiddleware(route.RefreshTokenAPI, client))

	//router for user
	router.POST("/api/content/image", middleware.AuthMiddleware(route.CreateContentImageAPI))

	//router for book API
	router.GET("/api/newest", middleware.AuthMiddleware(route.GetListNewestBookHeader))

	router.GET("/api/category/:category-id", middleware.AuthMiddleware(route.GetListCategoryBook))
	router.GET("/api/author/:author-id", middleware.AuthMiddleware(route.GetListAuthorBook))
	router.GET("/api/publisher/:publisher-id", middleware.AuthMiddleware(route.GetListPublisherBook))

	router.GET("/api/book/:id", middleware.AuthMiddleware(route.GetBookbyID))

	// Use handler of cors library to wrap the defined router above
	handler := cors.Default().Handler(router)
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000" // Default port if not specified
	}

	log.Println("Starting server listening at port ", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), handler))
}
