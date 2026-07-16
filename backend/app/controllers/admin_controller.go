package controllers

import (
	"fmt"
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gottatouchsomegrass/url/app/models"
	"github.com/gottatouchsomegrass/url/app/queries"
)

type AdminController struct {
	UserQuery *queries.UserQuery
	URLQuery  *queries.URLQuery
}

func NewAdminController(uq *queries.UserQuery, urlq *queries.URLQuery) *AdminController {
	return &AdminController{
		UserQuery: uq,
		URLQuery:  urlq,
	}
}

// GetAllUsers godoc
// @Summary      Get all users
// @Description  Get a paginated list of all users in the system (Admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page    query     int  false  "Page number" default(1)
// @Param        limit   query     int  false  "Items per page" default(50)
// @Success      200  {object}  models.PaginatedUserResponse
// @Failure      500  {object}  models.HTTPResponseErr
// @Router       /admin/users [get]
func (ac *AdminController) GetAllUsers(c *gin.Context) {
	ctx := c.Request.Context()

	limitStr := c.Query("limit")
	limit := 50 // Default value
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	pageStr := c.Query("page")
	if pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil && page > 1 {
			offset = (page - 1) * limit
		}
	}

	users, err := ac.UserQuery.GetAllUsers(ctx, limit, offset)
	if err != nil {
		fmt.Println("error getting users: \n", err)
		c.JSON(500, models.HTTPResponseErr{
			Err: "internal server error",
		})
		return
	}

	total, err := ac.UserQuery.CountAllUsers(ctx)
	if err != nil {
		fmt.Println("error counting users: \n", err)
		c.JSON(500, models.HTTPResponseErr{
			Err: "internal server error",
		})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	currentPage := (offset / limit) + 1

	c.JSON(200, models.PaginatedUserResponse{
		Data: users,
		Meta: models.PaginationMeta{
			Total:      total,
			Page:       currentPage,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

// GetUserURLsAsAdmin godoc
// @Summary      Get URLs of a specific user
// @Description  Get a paginated list of URLs belonging to a specific user (Admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        userid  path      int  true  "User ID"
// @Param        page    query     int  false  "Page number" default(1)
// @Param        limit   query     int  false  "Items per page" default(50)
// @Success      200  {object}  models.PaginatedURLResponse
// @Failure      400  {object}  models.HTTPError
// @Failure      500  {object}  models.HTTPResponseErr
// @Router       /admin/users/{userid}/urls [get]
func (ac *AdminController) GetUserURLsAsAdmin(c *gin.Context) {
	ctx := c.Request.Context()

	userIDStr := c.Param("userid")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(400, models.HTTPError{Error: "invalid user id"})
		return
	}

	limitStr := c.Query("limit")
	limit := 50 // Default value
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	pageStr := c.Query("page")
	if pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil && page > 1 {
			offset = (page - 1) * limit
		}
	}

	URLs, err := ac.URLQuery.GetUserURLs(ctx, userID, limit, offset)
	if err != nil {
		fmt.Println("error getting user urls: \n", err)
		c.JSON(500, models.HTTPResponseErr{
			Err: "internal server error",
		})
		return
	}

	total, err := ac.URLQuery.CountUserURLs(ctx, userID)
	if err != nil {
		fmt.Println("error counting user urls: \n", err)
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
