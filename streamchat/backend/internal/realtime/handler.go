package realtime

import (
	"context"
	"net/http"

	"time"

	"streamchat/backend/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	Hub           *Hub
	DB            *pgxpool.Pool
	Secret        string
	AllowedOrigin string
}

func (h *Handler) Register(r gin.IRoutes) { r.GET("/ws", h.serve) }
func (h *Handler) serve(c *gin.Context) {
	token, err := c.Cookie("streamchat_session")
	if err != nil {
		token = c.Query("access_token")
	}
	userID, err := auth.ParseToken(token, h.Secret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	origin := c.GetHeader("Origin")
	if origin != "" && origin != h.AllowedOrigin {
		c.JSON(http.StatusForbidden, gin.H{"error": "origin not allowed"})
		return
	}
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "" || r.Header.Get("Origin") == h.AllowedOrigin
	}}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	client, ok := h.Hub.Add(userID, conn)
	if !ok {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "connection limit reached"))
		_ = conn.Close()
		return
	}
	done := make(chan struct{})
	go client.WritePump(done)
	client.ReadPump()
	<-done
}
func NewMembershipChecker(db *pgxpool.Pool) func(string, string) bool {
	return func(userID, conversationID string) bool {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var ok bool
		return db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM conversation_members WHERE user_id=$1 AND conversation_id=$2)`, userID, conversationID).Scan(&ok) == nil && ok
	}
}
