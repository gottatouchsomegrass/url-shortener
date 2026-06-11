package controllers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gottatouchsomegrass/url/app/models"
	"github.com/gottatouchsomegrass/url/app/queries"
	"github.com/gottatouchsomegrass/url/pkg/utils"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	Query *queries.UserQuery
}

func NewAuthController(
	q *queries.UserQuery,
) *AuthController {
	return &AuthController{
		Query: q,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (auc *AuthController) Register(c *gin.Context) {
	ctx := c.Request.Context()
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	user := models.User{
		Email:        req.Email,
		PasswordHash: string(hash),
	}

	err = auc.Query.CreateUser(
		ctx,
		&user,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	token, err := utils.GenerateJWT(
		user.ID,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"token": token,
	})
}

func (auc *AuthController) Login(c *gin.Context) {
	ctx := c.Request.Context()
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user, err := auc.Query.GetUserByEmail(ctx, req.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(401, gin.H{
			"error": "invalid credentials",
		})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	token, err := utils.GenerateJWT(
		user.ID,
	)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "user logged in successfully",
		"token":   token,
	})
}

// Logout logs out the user by invalidating their token.(use then when using refresh tokens or Redis Blacklist)
func (auc *AuthController) Logout(c *gin.Context) {
	userID := c.MustGet(
		"userID",
	).(int64)
	if err := auc.Query.InvalidateToken(c.Request.Context(), userID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "user logged out successfully"})
}

// Me returns the user's information.
func (auc *AuthController) Me(
	c *gin.Context,
) {
	userID := c.MustGet(
		"userID",
	).(int64)

	user, err := auc.Query.GetUserByID(
		c.Request.Context(),
		userID,
	)

	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, user)
}
