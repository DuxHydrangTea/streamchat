package conversations

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Handler struct{ Service *Service }
type createRequest struct {
	Type            string   `json:"type" binding:"required"`
	Name            string   `json:"name"`
	MemberIDs       []string `json:"member_ids"`
	MemberUsernames []string `json:"member_usernames"`
}

func (h *Handler) Register(r gin.IRoutes) {
	r.GET("/conversations", h.list)
	r.POST("/conversations", h.create)
	r.GET("/conversations/:id/members", h.members)
}
func (h *Handler) list(c *gin.Context) {
	items, err := h.Service.List(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load conversations"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"conversations": items})
}
func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type is required"})
		return
	}
	item, err := h.Service.Create(c.Request.Context(), c.GetString("user_id"), req.Type, req.Name, req.MemberIDs, req.MemberUsernames)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}
func (h *Handler) members(c *gin.Context) {
	items, err := h.Service.Members(c.Request.Context(), c.GetString("user_id"), c.Param("id"))
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load members"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"members": items})
}
