package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gottatouchsomegrass/url/app/models"
	"github.com/gottatouchsomegrass/url/app/queries"
)

type AnalyticsController struct {
	Query *queries.AnalyticsQuery
}

func NewAnalyticsController(q *queries.AnalyticsQuery) *AnalyticsController {
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
		c.JSON(400, models.HTTPError{
			Error: "invalid url id",
		})
		return 0, false
	}

	url, err := ac.Query.GetURLByID(
		c.Request.Context(),
		urlID,
	)

	if err != nil {
		c.JSON(500, models.HTTPDetailsError{
			Error:   "internal server error",
			Details: err.Error(),
		})
		return 0, false
	}
	if url == nil {
		c.JSON(http.StatusNotFound, models.HTTPError{
			Error: "url not found",
		})
		return 0, false
	}

	if url.UserID != userID {
		c.JSON(403, models.HTTPError{
			Error: "forbidden",
		})
		return 0, false
	}

	return urlID, true
}

// GetAnalyticsOverview godoc
// @Summary      Get URL analytics overview
// @Description  Get click stats summary (today, this week, this month, total clicks) for a URL ID
// @Tags         analytics
// @Produce      json
// @Security     Bearer
// @Param        id path int true "URL ID"
// @Success      200  {object}  models.AnalyticsOverview
// @Failure      400  {object}  models.HTTPError
// @Failure      403  {object}  models.HTTPError
// @Failure      404  {object}  models.HTTPError
// @Failure      500  {object}  models.HTTPDetailsError
// @Router       /analytics/{id}/overview [get]
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
		c.JSON(http.StatusInternalServerError, models.HTTPDetailsError{
			Error:   "failed to fetch analytics",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// GetDailyClicks godoc
// @Summary      Get daily click analytics
// @Description  Get a list of dates and click counts for a URL ID
// @Tags         analytics
// @Produce      json
// @Security     Bearer
// @Param        id path int true "URL ID"
// @Success      200  {array}   models.DailyClick
// @Failure      400  {object}  models.HTTPError
// @Failure      403  {object}  models.HTTPError
// @Failure      404  {object}  models.HTTPError
// @Failure      500  {object}  models.HTTPDetailsError
// @Router       /analytics/{id}/daily [get]
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
		c.JSON(500, models.HTTPDetailsError{
			Error:   "failed to fetch clicks",
			Details: err.Error(),
		})
		return
	}

	c.JSON(200, clicks)
}

// GetRecentVisits godoc
// @Summary      Get recent visits logs
// @Description  Get up to 20 recent visitor events (IP, User Agent, Referer, Browser, Device, CreatedAt) for a URL ID
// @Tags         analytics
// @Produce      json
// @Security     Bearer
// @Param        id path int true "URL ID"
// @Success      200  {array}   models.RecentVisit
// @Failure      400  {object}  models.HTTPError
// @Failure      403  {object}  models.HTTPError
// @Failure      404  {object}  models.HTTPError
// @Failure      500  {object}  models.HTTPDetailsError
// @Router       /analytics/{id}/recent [get]
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
		c.JSON(http.StatusInternalServerError, models.HTTPDetailsError{
			Error:   "failed to fetch recent visits",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, visits)
}

// GetBrowserAnalytics godoc
// @Summary      Get browser distribution analytics
// @Description  Get visitor distribution stats grouped by browser type for a URL ID
// @Tags         analytics
// @Produce      json
// @Security     Bearer
// @Param        id path int true "URL ID"
// @Success      200  {array}   models.BrowserAnalytics
// @Failure      400  {object}  models.HTTPError
// @Failure      403  {object}  models.HTTPError
// @Failure      404  {object}  models.HTTPError
// @Failure      500  {object}  models.HTTPDetailsError
// @Router       /analytics/{id}/browser [get]
func (ac *AnalyticsController) GetBrowserAnalytics(
	c *gin.Context,
) {
	ctx := c.Request.Context()
	urlID, ok := ac.authorizeURL(c)
	if !ok {
		return
	}

	browserStats, err := ac.Query.GetBrowserAnalytics(
		ctx,
		urlID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.HTTPDetailsError{
			Error:   "failed to fetch browser analytics",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, browserStats)
}

// GetDeviceAnalytics godoc
// @Summary      Get device distribution analytics
// @Description  Get visitor distribution stats grouped by device type for a URL ID
// @Tags         analytics
// @Produce      json
// @Security     Bearer
// @Param        id path int true "URL ID"
// @Success      200  {array}   models.DeviceAnalytics
// @Failure      400  {object}  models.HTTPError
// @Failure      403  {object}  models.HTTPError
// @Failure      404  {object}  models.HTTPError
// @Failure      500  {object}  models.HTTPDetailsError
// @Router       /analytics/{id}/device [get]
func (ac *AnalyticsController) GetDeviceAnalytics(
	c *gin.Context,
) {
	ctx := c.Request.Context()
	urlID, ok := ac.authorizeURL(c)
	if !ok {
		return
	}

	deviceStats, err := ac.Query.GetDeviceAnalytics(
		ctx,
		urlID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.HTTPDetailsError{
			Error:   "failed to fetch device analytics",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, deviceStats)
}
