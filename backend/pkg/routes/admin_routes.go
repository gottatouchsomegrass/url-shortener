package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gottatouchsomegrass/url/app/controllers"
	"github.com/gottatouchsomegrass/url/pkg/middleware"
)

func AdminRoutes(r *gin.RouterGroup, ac *controllers.AdminController) {
	admin := r.Group("/admin")

	// Apply auth and require "admin" role
	admin.Use(middleware.AuthMiddleware(ac.UserQuery))
	admin.Use(middleware.RequireRole("admin"))

	// Admin endpoints
	admin.GET("/users",
		middleware.RateLimiter(ac.URLQuery.RDB, 100, time.Minute),
		ac.GetAllUsers,
	)

	admin.GET("/users/:userid/urls",
		middleware.RateLimiter(ac.URLQuery.RDB, 100, time.Minute),
		ac.GetUserURLsAsAdmin,
	)
}
