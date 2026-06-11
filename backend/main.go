package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gottatouchsomegrass/url/app/controllers"
	"github.com/gottatouchsomegrass/url/pkg/configs"
	"github.com/gottatouchsomegrass/url/pkg/middleware"
	"github.com/gottatouchsomegrass/url/pkg/routes"
	"github.com/gottatouchsomegrass/url/pkg/utils"
	"github.com/gottatouchsomegrass/url/platform/database"
	"github.com/joho/godotenv"
)

func main() {
	utils.InitValidator()

	// Load environment variables from .env files if present.
	// In cloud production environments, environment variables are typically injected directly.
	_ = godotenv.Load(".env.test")
	_ = godotenv.Load()

	r := gin.Default()

	// Enable CORS for cross-origin frontend requests
	r.Use(middleware.CORS())

	dbq, err := database.OpenDBConnection()
	if err != nil {
		log.Fatal(err)
	}

	//use global rate limiter
	// r.Use(
	//  middleware.RateLimiter(
	// 	dbq.RDB,
	// 	100,
	// 	time.Minute,
	//  ),
	// )

	urlController := controllers.NewURLController(
		dbq.URLQuery,
	)
	analyticsController := controllers.NewAnalyticsController(
		dbq.URLQuery,
	)
	authController := controllers.NewAuthController(
		dbq.UserQuery,
	)

	api := r.Group("/api/v1")
	routes.PublicRoutes(api, urlController, analyticsController, authController)
	routes.PrivateRoutes(api, urlController, analyticsController, authController)

	svr := configs.ConfigHTTPServer(r)
	utils.StartSvrGracefulShutdown(svr)
}
