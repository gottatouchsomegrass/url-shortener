// Package controllers handles HTTP request logic
package controllers

import (
	"fmt"
	"math"
	"strconv"
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

// ShortenURL godoc
// @Summary      Create a shortened URL
// @Description  Shorten a long URL, optionally with a custom code and expiry
// @Tags         urls
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body Request true "Shorten URL request"
// @Success      201  {object}  models.URLShortenSuccess
// @Failure      400  {object}  models.HTTPValidationError
// @Failure      500  {object}  models.HTTPError
// @Router       /shorten [post]
func (uc *URLController) ShortenURL(c *gin.Context) {
	ctx := c.Request.Context()

	var req Request

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, models.HTTPValidationError{
			Error: utils.ValidateErrors(err),
		})
		return
	}

	fmt.Println(req.LongURL)

	expiry := time.Now().Add(
		24 * time.Hour,
	)

	shortened := req.CustomCode

	userID := c.MustGet("userID").(int64)
	roleRaw, exists := c.Get("userRole")
	userRole := "free"
	if exists {
		userRole = roleRaw.(string)
	}

	// 1. Enforce Custom Code Limits
	if shortened != "" && userRole != "premium" && userRole != "admin" {
		c.JSON(403, models.HTTPError{
			Error: "Custom aliases require a premium subscription.",
		})
		return
	}

	// 2. Enforce Total URL Limits
	totalURLs, err := uc.Query.CountUserURLs(ctx, userID)
	if err != nil {
		c.JSON(500, models.HTTPError{Error: "Failed to check url limits"})
		return
	}

	if userRole == "free" && totalURLs >= 10 {
		c.JSON(403, models.HTTPError{
			Error: "Free plan limit reached (10 URLs). Please upgrade to base or premium.",
		})
		return
	}

	if userRole == "base" && totalURLs >= 1000 {
		c.JSON(403, models.HTTPError{
			Error: "Base plan limit reached (1000 URLs). Please upgrade to premium.",
		})
		return
	}

	if shortened == "" {
		shortened = utils.GenerateShortCode()
	}

	exist, err := uc.Query.CustomCodeExists(
		ctx,
		shortened,
	)

	if err != nil {
		c.JSON(500, models.HTTPError{
			Error: "failed checking shortcode",
		})
		return
	}

	if exist {
		c.JSON(400, models.HTTPError{
			Error: "custom code already exists",
		})
		return
	}

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
		c.JSON(500, models.HTTPError{
			Error: "Failed to create url",
		})
		return
	}
	c.JSON(201, models.URLShortenSuccess{
		Success: "url inserted to db",
	})
}

// RedirectURL godoc
// @Summary      Redirect short code
// @Description  Redirect a short code to its original long URL and log click analytics
// @Tags         urls
// @Param        shortCode path string true "Short URL Code"
// @Success      302  "Redirects to long URL"
// @Failure      404  {object}  models.HTTPResponseErr
// @Failure      410  {object}  models.HTTPResponseErr
// @Failure      500  {object}  models.HTTPResponseErr
// @Router       /{shortCode} [get]
func (uc *URLController) RedirectURL(c *gin.Context) {
	ctx := c.Request.Context()

	code := c.Param("shortCode")

	URL, err := uc.Query.GetByShortURL(ctx, code)
	if err != nil {
		fmt.Println("error shortcode: \n", err)
		c.JSON(500, models.HTTPResponseErr{
			Err: "internal server error",
		})
		return
	}

	if URL == nil {
		c.JSON(404, models.HTTPResponseErr{
			Err: "not found",
		})
		return
	}

	// expiry check
	if URL.Expiry != nil && time.Now().After(*URL.Expiry) {
		c.JSON(410, models.HTTPResponseErr{
			Err: "link expired",
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

// GetUserURLs godoc
// @Summary      Get user's shortened URLs
// @Description  Get a paginated list of URLs shortened by the authenticated user
// @Tags         urls
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page    query     int  false  "Page number" default(1)
// @Param        limit   query     int  false  "Items per page" default(50)
// @Success      200  {object}  models.PaginatedURLResponse
// @Failure      500  {object}  models.HTTPResponseErr
// @Router       /urls [get]
func (uc *URLController) GetUserURLs(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.MustGet("userID").(int64)

	// Simple pagination defaults
	// limit := 50
	// offset := 0

	limitStr := c.Query("limit")
	limit := 50 // Default value
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	} else {
		limit = 50
	}

	offset := 0
	pageStr := c.Query("page")
	if pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil && page > 1 {
			offset = (page - 1) * limit
		}
	} else {
		offset = 0
	}

	URLs, err := uc.Query.GetUserURLs(ctx, userID, limit, offset)
	if err != nil {
		fmt.Println("error getting user urls: \n", err)
		c.JSON(500, models.HTTPResponseErr{
			Err: "internal server error",
		})
		return
	}

	total, err := uc.Query.CountUserURLs(ctx, userID)
	if err != nil {
		fmt.Println("error getting user urls count: \n", err)
		c.JSON(500, models.HTTPResponseErr{
			Err: "internal server error",
		})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	currentPage := (offset / limit) + 1

	c.JSON(200, models.PaginatedURLResponse{
		Data: URLs,
		Meta: models.PaginationMeta{
			Total:      total,
			Page:       currentPage,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

type UpdateURLRequest struct {
	LongURL string `json:"long_url" binding:"required,url"`
}

// UpdateUserURL godoc
// @Summary      Update a shortened URL
// @Description  Update the long URL destination of an existing shortened URL
// @Tags         urls
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id      path      int  true  "URL ID"
// @Param        request body UpdateURLRequest true "Update URL request"
// @Success      200  {object}  models.MessageSuccess
// @Failure      400  {object}  models.HTTPValidationError
// @Failure      500  {object}  models.HTTPError
// @Router       /urls/{id} [put]
func (uc *URLController) UpdateUserURL(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.MustGet("userID").(int64)

	var req UpdateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, models.HTTPValidationError{
			Error: utils.ValidateErrors(err),
		})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, models.HTTPError{Error: "invalid url id"})
		return
	}

	err = uc.Query.UpdateURL(ctx, id, userID, req.LongURL)
	if err != nil {
		c.JSON(500, models.HTTPError{Error: err.Error()})
		return
	}

	c.JSON(200, models.MessageSuccess{Message: "url updated successfully"})
}

// DeleteUserURL godoc
// @Summary      Delete a shortened URL
// @Description  Delete a shortened URL by its ID
// @Tags         urls
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id      path      int  true  "URL ID"
// @Success      200  {object}  models.MessageSuccess
// @Failure      400  {object}  models.HTTPError
// @Failure      500  {object}  models.HTTPResponseErr
// @Router       /urls/{id} [delete]
func (uc *URLController) DeleteUserURL(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.MustGet("userID").(int64)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, models.HTTPError{Error: "invalid url id"})
		return
	}

	if err := uc.Query.DeleteURL(ctx, id, userID); err != nil {
		fmt.Println("error deleting url: \n", err)
		c.JSON(500, models.HTTPResponseErr{
			Err: "internal server error",
		})
		return
	}

	c.JSON(200, models.MessageSuccess{
		Message: "url deleted",
	})
}

type BulkDeleteRequest struct {
	IDs []int64 `json:"ids" binding:"required,min=1"`
}

// BulkDeleteUserURLs godoc
// @Summary      Bulk delete shortened URLs
// @Description  Delete multiple shortened URLs by their IDs
// @Tags         urls
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body BulkDeleteRequest true "Bulk delete request"
// @Success      200  {object}  models.MessageSuccess
// @Failure      400  {object}  models.HTTPValidationError
// @Failure      500  {object}  models.HTTPResponseErr
// @Router       /urls/bulk [delete]
func (uc *URLController) BulkDeleteUserURLs(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.MustGet("userID").(int64)

	var req BulkDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, models.HTTPValidationError{
			Error: utils.ValidateErrors(err),
		})
		return
	}

	if err := uc.Query.BulkDeleteURLs(ctx, req.IDs, userID); err != nil {
		fmt.Println("error bulk deleting urls: \n", err)
		c.JSON(500, models.HTTPResponseErr{
			Err: "failed to delete urls or unauthorized",
		})
		return
	}

	c.JSON(200, models.MessageSuccess{
		Message: fmt.Sprintf("successfully deleted %d urls", len(req.IDs)),
	})
}
