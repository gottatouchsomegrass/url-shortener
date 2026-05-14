package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/gottatouchsomegrass/url/app/controllers"
)

func PublicRoutes(r *gin.RouterGroup, uc *controllers.URLController) {
	r.GET("/:shortCode",uc.RedirectUrl)
	r.POST("/shorten",uc.ShortenURL)
}
