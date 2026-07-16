package controllers

import (
	"errors"
	"net/http"
	"os"
	"time"

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

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account and return a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Registration Info"
// @Success      201  {object}  models.AuthRegisterSuccess
// @Failure      400  {object}  models.HTTPError
// @Failure      500  {object}  models.HTTPError
// @Router       /auth/register [post]
func (auc *AuthController) Register(c *gin.Context) {
	ctx := c.Request.Context()
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, models.HTTPError{Error: err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		c.JSON(500, models.HTTPError{Error: err.Error()})
		return
	}

	tx, err := auc.Query.DB.Begin(ctx)
	if err != nil {
		c.JSON(500, models.HTTPError{Error: "Failed to begin transaction"})
		return
	}
	defer tx.Rollback(ctx)

	// 1. Create User
	user := models.User{
		Email:        req.Email,
		PasswordHash: string(hash),
	}
	err = auc.Query.CreateUserTx(ctx, tx, &user)
	if err != nil {
		c.JSON(500, models.HTTPError{Error: err.Error()})
		return
	}

	// 2. Create Refresh Session
	refresh, err := utils.GenerateRefreshToken()
	if err != nil {
		c.JSON(500, models.HTTPError{Error: err.Error()})
		return
	}
	refreshHash := utils.HashToken(refresh)

	rt := models.RefreshToken{
		UserID:      user.ID,
		RefreshHash: refreshHash,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		IP:          c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
	}
	err = auc.Query.CreateRefreshTokenTx(ctx, tx, &rt)
	if err != nil {
		c.JSON(500, models.HTTPError{Error: "Failed to store refresh token"})
		return
	}

	err = auc.Query.EnforceMaxRefreshTokensTx(ctx, tx, user.ID, 10)
	if err != nil {
		c.JSON(500, models.HTTPError{Error: "Failed to enforce max refresh sessions"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(500, models.HTTPError{Error: "Failed to commit transaction"})
		return
	}

	// 3. Generate JWT
	token, err := utils.GenerateJWT(
		user.ID,
		user.Role,
	)
	if err != nil {
		c.JSON(500, models.HTTPError{Error: err.Error()})
		return
	}

	// 4. Set Cookie
	secure := os.Getenv("APP_ENV") == "production"
	http.SetCookie(
		c.Writer,
		&http.Cookie{
			Name:     "refresh_token",
			Value:    refresh,
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
			Path:     "/api/v1/auth/refresh",
			MaxAge:   7 * 24 * 3600,
		},
	)

	// 5. Return Response
	c.JSON(201, models.AuthRegisterSuccess{
		Token: token,
	})
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate user and return a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login Info"
// @Success      200  {object}  models.AuthLoginSuccess
// @Failure      400  {object}  models.HTTPError
// @Failure      401  {object}  models.HTTPError
// @Failure      500  {object}  models.HTTPError
// @Router       /auth/login [post]
func (auc *AuthController) Login(c *gin.Context) {
	ctx := c.Request.Context()
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, models.HTTPError{Error: err.Error()})
		return
	}

	user, err := auc.Query.GetUserByEmail(ctx, req.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(401, models.HTTPError{
			Error: "invalid credentials",
		})
		return
	}
	if err != nil {
		c.JSON(500, models.HTTPDetailsError{Error: "user login db error", Details: err.Error()})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(401, models.HTTPError{Error: "invalid credentials"})
		return
	}

	// 2. Create Refresh Session
	tx, err := auc.Query.DB.Begin(ctx)
	if err != nil {
		c.JSON(500, models.HTTPError{Error: "Failed to begin transaction"})
		return
	}
	defer tx.Rollback(ctx)

	refresh, err := utils.GenerateRefreshToken()
	if err != nil {
		c.JSON(500, models.HTTPError{Error: err.Error()})
		return
	}

	rt := models.RefreshToken{
		UserID:      user.ID,
		RefreshHash: utils.HashToken(refresh),
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		IP:          c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
	}
	err = auc.Query.CreateRefreshTokenTx(ctx, tx, &rt)
	if err != nil {
		c.JSON(500, models.HTTPError{Error: "Failed to store refresh token"})
		return
	}

	err = auc.Query.EnforceMaxRefreshTokensTx(ctx, tx, user.ID, 10)
	if err != nil {
		c.JSON(500, models.HTTPError{Error: "Failed to enforce max refresh sessions"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(500, models.HTTPError{Error: "Failed to commit transaction"})
		return
	}

	// 3. Generate JWT
	token, err := utils.GenerateJWT(
		user.ID,
		user.Role,
	)
	if err != nil {
		c.JSON(500, models.HTTPError{
			Error: err.Error(),
		})
		return
	}

	// 4. Set Cookie
	secure := os.Getenv("APP_ENV") == "production"
	http.SetCookie(
		c.Writer,
		&http.Cookie{
			Name:     "refresh_token",
			Value:    refresh,
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
			Path:     "/api/v1/auth/refresh",
			MaxAge:   7 * 24 * 3600,
		},
	)

	// 5. Return Response
	c.JSON(200, models.AuthLoginSuccess{
		Message: "user logged in successfully",
		Token:   token,
	})
}

// Logout logs out the user by invalidating their token in the Redis Blacklist.
// @Summary      Logout user
// @Description  Invalidate the user's JWT token
// @Tags         auth
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  models.MessageSuccess
// @Failure      500  {object}  models.HTTPError
// @Router       /auth/logout [post]
func (auc *AuthController) Logout(c *gin.Context) {
	// 1. Get the token from the header (AuthMiddleware already verified it's present and valid)
	authHeader := c.GetHeader("Authorization")
	tokenString := authHeader[len("Bearer "):] // safe because of AuthMiddleware

	// 2. We should ideally parse the token to get its exact remaining expiration time,
	// but for simplicity, we can just blacklist it for the max JWT lifetime (24 hours).
	// An even better way is to pass the remaining time here.
	err := auc.Query.InvalidateToken(c.Request.Context(), tokenString, 24*time.Hour)
	if err != nil {
		c.JSON(500, models.HTTPError{Error: "failed to logout"})
		return
	}

	// Also clear the refresh token cookie
	secure := os.Getenv("APP_ENV") == "production"
	http.SetCookie(
		c.Writer,
		&http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
			Path:     "/api/v1/auth/refresh",
			MaxAge:   -1,
		},
	)
	// And optionally revoke it in DB if passed, but typically we just clear the cookie

	c.JSON(200, models.MessageSuccess{Message: "user logged out successfully"})
}

// Me returns the user's information.
// Me godoc
// @Summary      Get current user info
// @Description  Get current authenticated user info
// @Tags         auth
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  models.User
// @Failure      500  {object}  models.HTTPError
// @Router       /auth/me [get]
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
		c.JSON(500, models.HTTPError{
			Error: err.Error(),
		})
		return
	}

	c.JSON(200, user)
}

// DeleteAccount godoc
// @Summary      Delete user account
// @Description  Completely delete the user's account and all associated URLs and analytics
// @Tags         auth
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  models.MessageSuccess
// @Failure      500  {object}  models.HTTPError
// @Router       /auth/me [delete]
func (auc *AuthController) DeleteAccount(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	err := auc.Query.DeleteUserAccount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, models.HTTPError{
			Error: "Failed to delete account: " + err.Error(),
		})
		return
	}

	c.JSON(200, models.MessageSuccess{
		Message: "account deleted successfully",
	})
}

func (auc *AuthController) Refresh(c *gin.Context) {
	ctx := c.Request.Context()

	// Read refresh cookie
	refresh, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.HTTPError{
			Error: "missing refresh token",
		})
		return
	}

	hash := utils.HashToken(refresh)

	// Lookup refresh token
	rt, err := auc.Query.GetRefreshToken(ctx, hash)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.HTTPError{
			Error: "invalid refresh token",
		})
		return
	}

	// Check revoked
	if rt.RevokedAt != nil {
		c.JSON(http.StatusUnauthorized, models.HTTPError{
			Error: "refresh token revoked",
		})
		return
	}

	// Check expiry
	if time.Now().After(rt.ExpiresAt) {
		_ = auc.Query.DeleteRefreshToken(ctx, rt.ID)

		c.JSON(http.StatusUnauthorized, models.HTTPError{
			Error: "refresh token expired",
		})
		return
	}

	// Load user
	user, err := auc.Query.GetUserByID(ctx, rt.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.HTTPError{
			Error: "user not found",
		})
		return
	}

	// Generate new refresh token
	newRefresh, err := utils.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.HTTPError{
			Error: "failed generating refresh token",
		})
		return
	}

	newHash := utils.HashToken(newRefresh)

	// Rotate refresh token
	err = auc.Query.DeleteRefreshToken(ctx, rt.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.HTTPError{
			Error: "failed rotating refresh token",
		})
		return
	}

	newRT := models.RefreshToken{
		UserID:      user.ID,
		RefreshHash: newHash,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		IP:          c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
	}

	err = auc.Query.CreateRefreshToken(ctx, &newRT)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.HTTPError{
			Error: "failed creating refresh token",
		})
		return
	}

	// Generate new access token
	access, err := utils.GenerateJWT(
		user.ID,
		user.Role,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.HTTPError{
			Error: "failed generating access token",
		})
		return
	}

	// Send new refresh cookie
	c.SetCookie(
		"refresh_token",
		newRefresh,
		7*24*3600,
		"/",
		"",
		false, // secure=true in production
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"token": access,
	})
}
