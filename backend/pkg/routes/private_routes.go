package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gottatouchsomegrass/url/app/controllers"
	"github.com/gottatouchsomegrass/url/pkg/middleware"
)

func PrivateRoutes(r *gin.RouterGroup,
	uc *controllers.URLController,
	ac *controllers.AnalyticsController,
	auc *controllers.AuthController) {
	//auth private endpoints
	r.Use(
		middleware.AuthMiddleware(),
	)
	auth := r.Group("/auth")
	auth.POST("/logout", auc.Logout)
	auth.GET("/me", auc.Me)

	// analytics private endpoints
	analytics := r.Group("/analytics")
	analytics.GET(
		"/:id/overview",
		middleware.RateLimiter(
			uc.Query.RDB,
			100,
			time.Minute,
		),
		ac.GetAnalyticsOverview,
	)
	analytics.GET(
		"/:id/daily",
		middleware.RateLimiter(
			uc.Query.RDB,
			100,
			time.Minute,
		),
		ac.GetDailyClicks,
	)
	analytics.GET(
		"/:id/recent",
		middleware.RateLimiter(
			uc.Query.RDB,
			100,
			time.Minute,
		),
		ac.GetRecentVisits,
	)

	//url private endpoints
	r.POST("/shorten",
		middleware.RateLimiter(
			uc.Query.RDB,
			20,
			time.Minute,
		),
		uc.ShortenURL)

}
