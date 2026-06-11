// Package routes include all the routes
package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gottatouchsomegrass/url/app/controllers"
	"github.com/gottatouchsomegrass/url/pkg/middleware"
)

func PublicRoutes(r *gin.RouterGroup,
	uc *controllers.URLController,
	ac *controllers.AnalyticsController,
	authController *controllers.AuthController) {

	//auth public routes
	auth := r.Group("/auth")
	auth.POST("/register", authController.Register)
	auth.POST("/login", authController.Login)

	//url public routes
	r.GET("/:shortCode",
		middleware.RateLimiter(
			uc.Query.RDB,
			100,
			time.Minute,
		),
		uc.RedirectURL)
}
