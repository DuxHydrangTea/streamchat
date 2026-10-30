package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"streamchat/backend/internal/realtime"
	"streamchat/backend/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	Jobs    JobStore
	Queue   Queue
	Policy  URLPolicy
	Storage storage.Store
	Hub     *realtime.Hub
}
type createRequest struct {
	URL             string `json:"url" binding:"required"`
	ClientMessageID string `json:"client_message_id" binding:"required"`
}

func (h *Handler) Register(r gin.IRoutes) {
	r.POST("/conversations/:id/media-jobs", h.create)
	r.GET("/media/jobs/:id", h.get)
	r.GET("/media/jobs/:id/content", h.content)
	r.POST("/media/jobs/:id/retry", h.retry)
	r.POST("/media/jobs/:id/cancel", h.cancel)
}
func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url and client_message_id are required"})

		return
	}
	provider, source, err := h.Policy.Validate(c.Request.Context(), req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := uuid.Parse(req.ClientMessageID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_message_id must be a UUID"})
		return
	}
	job, messageID, created, err := h.Jobs.Create(c.Request.Context(), c.Param("id"), c.GetString("user_id"), req.ClientMessageID, source, provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not create video message"})
		return
	}
	if created {
		if err := EnqueueDownload(c.Request.Context(), h.Queue, job); err != nil {
			_ = h.Jobs.SetStatus(c.Request.Context(), job.ID, "failed", "enqueue_failed", "media job could not be queued")
			job.Status, job.ErrorCode, job.ErrorMessage = "failed", "enqueue_failed", "media job could not be queued"
		}
	}
	var senderName string
	if err := h.Jobs.DB.QueryRow(c.Request.Context(), `SELECT display_name FROM users WHERE id=$1`, c.GetString("user_id")).Scan(&senderName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load sender profile"})
		return
	}
	message := gin.H{"id": messageID, "client_message_id": req.ClientMessageID, "conversation_id": job.ConversationID, "sender_id": c.GetString("user_id"), "sender_name": senderName, "type": "video", "text_content": source, "created_at": time.Now().UTC(), "media_job": gin.H{"id": job.ID, "status": job.Status, "error_message": job.ErrorMessage}}

	if created && h.Hub != nil {
		h.Hub.Publish(job.ConversationID, realtime.Event{Type: "message.created", ConversationID: job.ConversationID, Payload: message})
	}
	c.JSON(http.StatusAccepted, message)

}
func (h *Handler) get(c *gin.Context) {
	job, err := h.getJob(c.Request.Context(), c.Param("id"), c.GetString("user_id"))
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "media job not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load media job"})
		return
	}
	c.JSON(http.StatusOK, job)
}
func (h *Handler) retry(c *gin.Context) {
	job, err := h.getJob(c.Request.Context(), c.Param("id"), c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "media job not found"})
		return
	}
	if job.Status != "failed" {
		c.JSON(http.StatusConflict, gin.H{"error": "only failed jobs can be retried"})
		return
	}
	if job.Attempts >= 3 {
		c.JSON(http.StatusConflict, gin.H{"error": "media job retry limit reached"})
		return
	}
	if err := h.Jobs.Retry(c.Request.Context(), job.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retry media job"})
		return
	}
	if err := enqueueDownload(c.Request.Context(), h.Queue, job.Job, job.ID+"-retry-"+strconv.Itoa(job.Attempts+1)); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "media job could not be queued"})
		return
	}
	job.Status = "queued"
	job.ErrorCode, job.ErrorMessage = "", ""
	c.JSON(http.StatusAccepted, job)
}
func (h *Handler) content(c *gin.Context) {
	var key, mime, filename string
	var size int64
	err := h.Jobs.DB.QueryRow(c.Request.Context(), `SELECT j.storage_key,j.mime_type,j.size_bytes,COALESCE(j.title,'video.mp4') FROM media_jobs j JOIN conversation_members cm ON cm.conversation_id=j.conversation_id WHERE j.id=$1 AND cm.user_id=$2 AND j.status='completed'`, c.Param("id"), c.GetString("user_id")).Scan(&key, &mime, &size, &filename)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not access video"})
		return
	}
	if h.Storage == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "video storage is unavailable"})
		return
	}
	obj, err := h.Storage.Open(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}
	defer obj.Close()
	if mime == "" {
		mime = "video/mp4"
	}
	filename = filepath.Base(filename)
	c.Header("Content-Type", mime)
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, strings.ReplaceAll(filename, `"`, "")))
	c.Header("Accept-Ranges", "bytes")
	start, end, partial, valid := parseRange(c.GetHeader("Range"), size)
	if !valid {
		c.Header("Content-Range", fmt.Sprintf("bytes */%d", size))
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return
	}
	if seeker, ok := obj.(io.Seeker); ok {
		if _, err := seeker.Seek(start, io.SeekStart); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		length := end - start + 1
		if partial {
			c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
			c.Status(http.StatusPartialContent)
		}
		c.Header("Content-Length", fmt.Sprint(length))
		_, _ = io.CopyN(c.Writer, obj, length)
		return
	}
	if partial {
		c.Header("Content-Range", fmt.Sprintf("bytes */%d", size))
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return
	}
	c.Header("Content-Length", fmt.Sprint(size))
	_, _ = io.CopyN(c.Writer, obj, size)
}

func parseRange(value string, size int64) (int64, int64, bool, bool) {
	if size <= 0 {
		return 0, 0, false, false
	}
	if value == "" {
		return 0, size - 1, false, size > 0
	}
	if !strings.HasPrefix(value, "bytes=") || strings.Contains(value, ",") {
		return 0, 0, false, false
	}
	parts := strings.SplitN(strings.TrimPrefix(value, "bytes="), "-", 2)
	if len(parts) != 2 || size < 1 {
		return 0, 0, false, false
	}
	if parts[0] == "" {
		suffix, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || suffix < 1 {
			return 0, 0, false, false
		}
		if suffix > size {
			suffix = size
		}
		return size - suffix, size - 1, true, true
	}
	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || start < 0 || start >= size {
		return 0, 0, false, false
	}
	end := size - 1
	if parts[1] != "" {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil || end < start {
			return 0, 0, false, false
		}
		if end >= size {
			end = size - 1
		}
	}
	return start, end, true, true
}

func (h *Handler) cancel(c *gin.Context) {
	job, err := h.getJob(c.Request.Context(), c.Param("id"), c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "media job not found"})
		return
	}
	if job.Status == "completed" || job.Status == "cancelled" {
		c.JSON(http.StatusConflict, gin.H{"error": "job cannot be cancelled"})
		return
	}
	if err := h.Jobs.SetStatus(c.Request.Context(), job.ID, "cancelled", "", ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not cancel media job"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

type storedJob struct{ Job }

func (h *Handler) getJob(ctx context.Context, id, userID string) (storedJob, error) {
	var j storedJob
	var duration *float64
	var width, height *int
	var size *int64
	err := h.Jobs.DB.QueryRow(ctx, `SELECT id::text,conversation_id::text,COALESCE(message_id::text,''),requested_by::text,source_url,provider::text,status::text,attempts,COALESCE(error_code,''),COALESCE(error_message,''),title,duration_seconds,width,height,size_bytes FROM media_jobs WHERE id=$1 AND requested_by=$2`, id, userID).Scan(&j.Job.ID, &j.Job.ConversationID, &j.Job.MessageID, &j.Job.RequestedBy, &j.Job.SourceURL, &j.Job.Provider, &j.Job.Status, &j.Job.Attempts, &j.Job.ErrorCode, &j.Job.ErrorMessage, &j.Job.Title, &duration, &width, &height, &size)
	j.Duration, j.Width, j.Height, j.Size = duration, width, height, size
	if j.Status == "completed" {
		j.PlaybackURL = "/api/v1/media/jobs/" + j.ID + "/content"
	}
	return j, err
}
