package messages

import (
	"errors"
	"net/http"
	"strconv"

	"streamchat/backend/internal/realtime"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	Service *Service
	Hub     *realtime.Hub
	DB      *pgxpool.Pool
}
type createRequest struct {
	ClientMessageID string `json:"client_message_id" binding:"required"`
	Text            string `json:"text_content" binding:"required"`
}

func (h *Handler) Register(r gin.IRoutes) {
	r.GET("/conversations/:id/messages", h.list)
	r.POST("/conversations/:id/messages", h.create)
	r.POST("/conversations/:id/read", h.read)
}
func (h *Handler) list(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	page, err := h.Service.List(c.Request.Context(), c.Param("id"), c.GetString("user_id"), c.Query("before"), limit)
	if errors.Is(err, ErrNotMember) {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not load messages"})
		return
	}
	c.JSON(http.StatusOK, page)
}
func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_message_id and text_content are required"})
		return
	}
	msg, err := h.Service.Create(c.Request.Context(), CreateInput{ConversationID: c.Param("id"), SenderID: c.GetString("user_id"), ClientMessageID: req.ClientMessageID, Text: req.Text})
	if errors.Is(err, ErrInvalidMessage) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, ErrNotMember) {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not create message"})
		return
	}
	h.Hub.Publish(msg.ConversationID, realtime.Event{Type: "message.created", ConversationID: msg.ConversationID, Payload: msg})
	c.JSON(http.StatusCreated, msg)
}
func (h *Handler) read(c *gin.Context) {
	var req struct {
		MessageID string `json:"message_id" binding:"required"`
	}
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message_id is required"})
		return
	}
	if err := h.Service.MarkRead(c.Request.Context(), c.Param("id"), c.GetString("user_id"), req.MessageID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "could not update read state"})
		}
		return
	}
	h.Hub.Publish(c.Param("id"), realtime.Event{Type: "message.read", ConversationID: c.Param("id"), Payload: gin.H{"user_id": c.GetString("user_id"), "message_id": req.MessageID}})
	c.Status(http.StatusNoContent)
}
