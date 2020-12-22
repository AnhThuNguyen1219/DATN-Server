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
	router.POST("/auth/signup", middleware.AddRedisClientMiddleware(route.SignUp, client))
	router.POST("/api/token/refresh", middleware.AddRedisClientMiddleware(route.RefreshTokenAPI, client))

	//router for book API
	router.GET("/api/newest", route.GetListNewestBookHeader)
	router.GET("/api/category", route.GetListCategoryName)
	router.GET("/api/category/:category-id", route.GetListCategoryBook)
	router.GET("/api/book", route.GetSearchBook)
	router.POST("/api/book", middleware.AuthMiddleware(route.PostANewBook))
	//author
	router.POST("/api/author", middleware.AuthMiddleware(route.PostANewAuthor))
	router.GET("/api/author", route.GetListAuthor)
	router.GET("/api/author/:author-id", route.GetListAuthorBook)
	//publisher
	router.GET("/api/publisher", route.GetListPublisher)
	router.GET("/api/publisher/:publisher-id", route.GetListPublisherBook)
	router.POST("/api/publisher", middleware.AuthMiddleware(route.PostANewPublisher))

	router.GET("/api/book/:id", route.GetBookbyID)
	router.DELETE("/api/book/:id", route.DelBookbyID)

	router.GET("/api/book/:id/review", route.GetListReviewofBook)
	router.GET("/api/book/:id/review-sum", route.GetSumReviewofBook)

	//router for user API
	// router.GET("/api/new/user", route.NewUser)
	router.GET("/api/user/:id", route.GetUserByID)
	router.GET("/api/user/:id/favourite", route.GetListFavourBookofUser)
	router.GET("/api/user/:id/review", route.GetListReviewofUser)
	//review
	router.GET("/api/review/:id", route.GetReview)
	router.PUT("/api/review/:id", middleware.AuthMiddleware(route.PutAReview))
	router.DELETE("/api/review/:id", middleware.AuthMiddleware(route.DelAReview))
	//list favourite

	router.POST("/api/favourite", middleware.AuthMiddleware(route.PostFavourABook))
	router.POST("/api/review", middleware.AuthMiddleware(route.PostReviewABook))

	// Use handler of cors library to wrap the defined router above
	handler := cors.AllowAll().Handler(router)
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000" // Default port if not specified
	}

	log.Println("Starting server listening at port ", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), handler))
}
