package attachments

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Handler struct{ Service *Service }

func (h *Handler) Register(r gin.IRoutes) {
	r.POST("/conversations/:id/attachments", h.upload)
	r.POST("/attachments/:id/complete", h.complete)
	r.POST("/attachments/:id/abort", h.abort)
	r.GET("/attachments/:id/access", h.access)
}
func (h *Handler) upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxUploadBytes+1)
	if err := c.Request.ParseMultipartForm(2 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart upload"})
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()
	if header.Size > MaxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file is too large"})
		return
	}
	mime := header.Header.Get("Content-Type")
	a, err := h.Service.Create(c.Request.Context(), c.GetString("user_id"), c.Param("id"), header.Filename, mime, header.Size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Service.Put(c.Request.Context(), a, io.LimitReader(file, MaxUploadBytes+1)); err != nil {
		_ = h.Service.Abort(c.Request.Context(), c.GetString("user_id"), a.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}
	a, err = h.Service.Complete(c.Request.Context(), c.GetString("user_id"), a.ID, header.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload verification failed"})
		return
	}
	c.JSON(http.StatusCreated, a)
}
func (h *Handler) complete(c *gin.Context) {
	size, _ := strconv.ParseInt(c.PostForm("size"), 10, 64)
	a, err := h.Service.Complete(c.Request.Context(), c.GetString("user_id"), c.Param("id"), size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "upload cannot be completed"})
		return
	}
	c.JSON(http.StatusOK, a)
}
func (h *Handler) abort(c *gin.Context) {
	if err := h.Service.Abort(c.Request.Context(), c.GetString("user_id"), c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "upload cannot be aborted"})
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) access(c *gin.Context) {
	a, err := h.Service.Access(c.Request.Context(), c.GetString("user_id"), c.Param("id"))
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "attachment not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not access attachment"})
		return
	}
	f, err := h.Service.Store.Open(c.Request.Context(), a.StorageKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
		return
	}
	defer f.Close()
	c.Header("Content-Disposition", "inline; filename=\""+a.Filename+"\"")
	c.DataFromReader(http.StatusOK, a.Size, a.MimeType, f, nil)
}
