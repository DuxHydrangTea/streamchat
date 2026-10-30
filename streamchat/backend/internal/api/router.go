package api

import (
	"context"
	"net"
	"net/http"
	"os"
	"streamchat/backend/internal/storage"
	"time"

	"streamchat/backend/internal/attachments"
	"streamchat/backend/internal/auth"
	"streamchat/backend/internal/conversations"
	"streamchat/backend/internal/media"
	"streamchat/backend/internal/messages"
	"streamchat/backend/internal/realtime"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	JWTSecret      string
	SecureCookie   bool
	FrontendOrigin string
}

func NewRouter(db *pgxpool.Pool, cfg Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), securityHeaders(), cors(cfg.FrontendOrigin), requestTimeout(10*time.Second))
	r.GET("/health/live", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/health/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	v1 := r.Group("/api/v1")
	authHandler := &auth.Handler{DB: db, Secret: cfg.JWTSecret, SecureCookie: cfg.SecureCookie}
	authHandler.Register(v1)
	protected := v1.Group("")
	protected.Use(authHandler.Middleware())
	hub := realtime.NewHub(realtime.NewMembershipChecker(db))
	(&conversations.Handler{Service: &conversations.Service{DB: db}}).Register(protected)
	(&messages.Handler{Service: &messages.Service{DB: db}, Hub: hub, DB: db}).Register(protected)
	(&attachments.Handler{Service: &attachments.Service{DB: db, Store: storage.Local{Root: envStorageRoot()}}}).Register(protected)
	redisAddr := os.Getenv("REDIS_ADDR")
	mediaQueue := media.Queue{Client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})}
	(&media.Handler{Jobs: media.JobStore{DB: db}, Queue: mediaQueue, Policy: media.URLPolicy{Resolver: net.DefaultResolver}, Storage: storage.Local{Root: envStorageRoot()}, Hub: hub}).Register(protected)
	(&realtime.Handler{Hub: hub, DB: db, Secret: cfg.JWTSecret, AllowedOrigin: cfg.FrontendOrigin}).Register(r)
	return r
}
func envStorageRoot() string {
	if root := os.Getenv("LOCAL_STORAGE_ROOT"); root != "" {
		return root
	}
	return "./data"
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
func cors(origin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Origin") == origin {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
func requestTimeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
