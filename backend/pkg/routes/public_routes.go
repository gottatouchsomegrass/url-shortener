//Package routes include all the routes
package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gottatouchsomegrass/url/app/controllers"
	"github.com/gottatouchsomegrass/url/pkg/middleware"
)

func PublicRoutes(r *gin.RouterGroup, uc *controllers.URLController) {
	r.GET("/:shortCode",
		middleware.RateLimiter(
			uc.Query.RDB,
			100,
			time.Minute,
		),
		uc.RedirectURL)
	r.POST("/shorten",
		middleware.RateLimiter(
			uc.Query.RDB,
			20,
			time.Minute,
		),
		uc.ShortenURL)
}
