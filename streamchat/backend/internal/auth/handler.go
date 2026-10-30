package auth

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

type Handler struct {
	DB           *pgxpool.Pool
	Secret       string
	SecureCookie bool
}
type credentials struct {
	Username    string `json:"username" binding:"required"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password" binding:"required"`
}
type userResponse struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

func (h *Handler) Register(r gin.IRoutes) {
	r.POST("/auth/register", h.register)
	r.POST("/auth/login", h.login)
	r.POST("/auth/logout", h.logout)
	r.GET("/auth/me", h.authenticate(), h.me)
}

func (h *Handler) register(c *gin.Context) {
	var req credentials
	if err := c.ShouldBindJSON(&req); err != nil || !usernamePattern.MatchString(req.Username) || len(req.Password) < 10 || len(req.Password) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username or password (password must be 10-72 bytes)"})
		return
	}
	req.Username = strings.ToLower(req.Username)
	if req.DisplayName == "" {
		req.DisplayName = req.Username
	}
	if len(req.DisplayName) > 80 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "display name is too long"})
		return
	}
	hash, err := HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create account"})
		return
	}
	var user userResponse
	err = h.DB.QueryRow(c.Request.Context(), `INSERT INTO users(username, display_name, password_hash) VALUES($1,$2,$3) RETURNING id::text, username, display_name`, req.Username, req.DisplayName, hash).Scan(&user.ID, &user.Username, &user.DisplayName)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username is unavailable"})
		return
	}
	h.setToken(c, user.ID)
	c.JSON(http.StatusCreated, user)
}

func (h *Handler) login(c *gin.Context) {
	var req credentials
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}
	var user userResponse
	var hash string
	err := h.DB.QueryRow(c.Request.Context(), `SELECT id::text, username, display_name, password_hash FROM users WHERE username=$1`, strings.ToLower(req.Username)).Scan(&user.ID, &user.Username, &user.DisplayName, &hash)
	if err != nil || !VerifyPassword(hash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	h.setToken(c, user.ID)
	c.JSON(http.StatusOK, user)
}

func (h *Handler) logout(c *gin.Context) {
	c.SetCookie("streamchat_session", "", -1, "/", "", h.SecureCookie, true)
	c.Status(http.StatusNoContent)
}
func (h *Handler) me(c *gin.Context) {
	var user userResponse
	err := h.DB.QueryRow(c.Request.Context(), `SELECT id::text, username, display_name FROM users WHERE id=$1`, c.GetString("user_id")).Scan(&user.ID, &user.Username, &user.DisplayName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}
func (h *Handler) setToken(c *gin.Context, userID string) {
	token, err := SignToken(userID, h.Secret, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create session"})
		return
	}
	c.SetCookie("streamchat_session", token, 900, "/", "", h.SecureCookie, true)
}
func (h *Handler) Middleware() gin.HandlerFunc {
	return h.authenticate()
}

func (h *Handler) authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("streamchat_session")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		id, err := ParseToken(token, h.Secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}
		c.Set("user_id", id)
		c.Next()
	}
}

var _ = pgx.ErrNoRows
