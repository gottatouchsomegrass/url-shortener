// Package controllers handles HTTP request logic
package controllers

import (
	"fmt"
	"time"

	"github.com/mileusna/useragent"

	"github.com/gin-gonic/gin"
	"github.com/gottatouchsomegrass/url/app/models"
	"github.com/gottatouchsomegrass/url/app/queries"
	"github.com/gottatouchsomegrass/url/pkg/utils"
)

type URLController struct {
	Query *queries.URLQuery
}

type Request struct {
	LongURL    string     `json:"long_url" binding:"required,url"`
	CustomCode string     `json:"custom_code,omitempty" binding:"omitempty,shortcode"`
	Expiry     *time.Time `json:"expiry,omitempty"`
}

// NewURLController adds new url to db
func NewURLController(q *queries.URLQuery) *URLController {
	return &URLController{Query: q}
}

func (uc *URLController) ShortenURL(c *gin.Context) {
	ctx := c.Request.Context()

	var req Request

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": utils.ValidateErrors(err),
		})
		return
	}

	fmt.Println(req.LongURL)

	expiry := time.Now().Add(
		24 * time.Hour,
	)

	shortened := req.CustomCode
	if shortened == "" {
		shortened = utils.GenerateShortCode()
	}

	exists, err := uc.Query.CustomCodeExists(
		ctx,
		shortened,
	)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed checking shortcode",
		})
		return
	}

	if exists {
		c.JSON(400, gin.H{
			"error": "custom code already exists",
		})
		return
	}

	userID := c.MustGet("userID").(int64)
	newURL := models.URL{
		UserID:    userID,
		LongURL:   req.LongURL,
		ShortURL:  shortened,
		Expiry:    &expiry,
		Clicks:    0,
		CreatedAt: time.Now(),
	}

	err = uc.Query.CreateURL(ctx, &newURL)
	if err != nil {
		fmt.Println("query err shorten: \n", err)
		c.JSON(500, gin.H{
			"error": "Failed to create url",
		})
		return
	}
	c.JSON(201, gin.H{
		"success": "url inserted to db",
	})
}

// RedirectURL redirects url
func (uc *URLController) RedirectURL(c *gin.Context) {
	ctx := c.Request.Context()

	code := c.Param("shortCode")

	URL, err := uc.Query.GetByShortURL(ctx, code)
	if err != nil {
		fmt.Println("error shortcode: \n", err)
		c.JSON(500, gin.H{
			"err": "internal server error",
		})
		return
	}

	if URL == nil {
		c.JSON(404, gin.H{
			"err": "not found",
		})
		return
	}

	// expiry check
	if URL.Expiry != nil && time.Now().After(*URL.Expiry) {
		c.JSON(410, gin.H{
			"err": "link expired",
		})
		return
	}

	//increment clicks
	if err := uc.Query.IncrementClicks(ctx, URL.ID); err != nil {
		fmt.Println("query err increment clicks: \n", err)
	}

	ua := useragent.Parse(c.Request.UserAgent())

	browser := ua.Name
	device := ua.Device

	// record analytics event
	event := models.ClickEvent{
		URLID:     URL.ID,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Referer:   c.Request.Referer(),
		Country:   "",
		Device:    device,
		Browser:   browser,
	}

	if err := uc.Query.CreateClickEvent(ctx, &event); err != nil {
		fmt.Println(err)
	}

	c.Redirect(302, URL.LongURL)
}
