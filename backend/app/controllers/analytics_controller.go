package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gottatouchsomegrass/url/app/queries"
)

type AnalyticsController struct {
	Query *queries.URLQuery
}

func NewAnalyticsController(q *queries.URLQuery) *AnalyticsController {
	return &AnalyticsController{
		Query: q,
	}
}

func (ac *AnalyticsController) authorizeURL(
	c *gin.Context,
) (int64, bool) {

	userID := c.MustGet("userID").(int64)

	urlID, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)

	if err != nil {
		c.JSON(400, gin.H{
			"error": "invalid url id",
		})
		return 0, false
	}

	url, err := ac.Query.GetURLByID(
		c.Request.Context(),
		urlID,
	)

	if err != nil {
		c.JSON(404, gin.H{
			"error": "url not found",
		})
		return 0, false
	}

	if url.UserID != userID {
		c.JSON(403, gin.H{
			"error": "forbidden",
		})
		return 0, false
	}

	return urlID, true
}

func (ac *AnalyticsController) GetAnalyticsOverview(c *gin.Context) {
	ctx := c.Request.Context()
	urlID, ok := ac.authorizeURL(c)
	if !ok {
		return
	}

	overview, err := ac.Query.GetAnalyticsOverview(
		ctx,
		urlID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch analytics",
		})
		return
	}

	c.JSON(http.StatusOK, overview)
}

func (ac *AnalyticsController) GetDailyClicks(
	c *gin.Context,
) {
	ctx := c.Request.Context()
	urlID, ok := ac.authorizeURL(c)
	if !ok {
		return
	}

	clicks, err := ac.Query.GetDailyClicks(
		ctx,
		urlID,
	)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to fetch clicks",
		})
		return
	}

	c.JSON(200, clicks)
}

func (ac *AnalyticsController) GetRecentVisits(
	c *gin.Context,
) {
	ctx := c.Request.Context()
	urlID, ok := ac.authorizeURL(c)
	if !ok {
		return
	}

	visits, err := ac.Query.GetRecentVisits(
		ctx,
		urlID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch recent visits",
		})
		return
	}

	c.JSON(http.StatusOK, visits)
}
