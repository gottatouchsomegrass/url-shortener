// Package controllers handles HTTP request logic
package controllers

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gottatouchsomegrass/url/app/models"
	"github.com/gottatouchsomegrass/url/app/queries"
	"github.com/gottatouchsomegrass/url/pkg/utils"
)

type URLController struct {
	Query *queries.URLQuery
}

type Request struct {
	LongURL string `json:"long_url" validate:"required,url"`
	CustomCode string `json:"custom_code,omitempty" validate:"omitempty,shortcode"`
}

func NewURLController(q *queries.URLQuery) *URLController {
	return &URLController{Query : q}
}

func (uc *URLController) ShortenURL(c *gin.Context) {
	ctx := c.Request.Context()
	
	var req Request

	if err:=c.ShouldBindJSON(&req); err!=nil {
		c.JSON(400, gin.H{
			"error" : "invalid req body",
		})
		return 
	}

	fmt.Println(req.LongURL)
	
	expiry := time.Now().Add(
		24 * time.Hour,
	)
	shortened := utils.GenerateShortCode()

	newURL := models.URL{
		LongURL: req.LongURL,
		ShortURL: shortened,
		Expiry: &expiry,
		Clicks: 0,
		CreatedAt: time.Now(),
	}

	err := uc.Query.CreateURL(ctx, &newURL)
	if err!=nil {
		fmt.Println("query err shorten: \n",err)
		c.JSON(500, gin.H{
			"error" : "Failed to create url",
		})
		return
	}
	c.JSON(201, gin.H{
		"success" : "url inserted to db",
	})
}

func (uc *URLController) RedirectURL(c *gin.Context) {
	ctx := c.Request.Context()
	
	code := c.Param("shortCode")

	URL,err := uc.Query.GetByShortURL(ctx,code)
	if err!=nil {
		fmt.Println("error shortcode: \n",err)
		c.JSON(500,gin.H{
			"err" : "internal server error",
		})
		return
	}

	if URL==nil {
		c.JSON(404,gin.H{
			"err" : "not found",
		})
		return
	}

	if URL.Expiry!=nil && time.Now().After(*URL.Expiry) {
		c.JSON(410, gin.H{
			"err" : "link expired",
		})
		return
	}

	c.Redirect(302,URL.LongURL)
}
