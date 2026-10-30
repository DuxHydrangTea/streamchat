package messages

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalidMessage = errors.New("message must contain 1-4000 characters")
var ErrNotMember = errors.New("conversation not found")

type MediaJob struct {
	ID           string  `json:"id"`
	Status       string  `json:"status"`
	ErrorMessage string  `json:"error_message,omitempty"`
	Title        string  `json:"title,omitempty"`
	Duration     float64 `json:"duration_seconds,omitempty"`
	Width        int     `json:"width,omitempty"`
	Height       int     `json:"height,omitempty"`
	Size         int64   `json:"size_bytes,omitempty"`
	PlaybackURL  string  `json:"playback_url,omitempty"`
}

type Message struct {
	ID              string    `json:"id"`
	ConversationID  string    `json:"conversation_id"`
	SenderID        string    `json:"sender_id"`
	SenderName      string    `json:"sender_name"`
	ClientMessageID string    `json:"client_message_id"`
	Type            string    `json:"type"`
	Text            string    `json:"text_content"`
	CreatedAt       time.Time `json:"created_at"`
	MediaJob        *MediaJob `json:"media_job,omitempty"`
}
type Page struct {
	Messages   []Message `json:"messages"`
	NextCursor string    `json:"next_cursor,omitempty"`
	HasMore    bool      `json:"has_more"`
}
type Service struct{ DB *pgxpool.Pool }

type CreateInput struct{ ConversationID, SenderID, ClientMessageID, Text string }

func ValidateText(text string) error {
	text = strings.TrimSpace(text)
	if len([]rune(text)) == 0 || len([]rune(text)) > 4000 {
		return ErrInvalidMessage
	}
	return nil
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Message, error) {
	if err := ValidateText(input.Text); err != nil {
		return Message{}, err
	}
	if _, err := uuid.Parse(input.ClientMessageID); err != nil {
		return Message{}, errors.New("client_message_id must be a UUID")
	}
	var member bool
	if err := s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM conversation_members WHERE conversation_id=$1 AND user_id=$2)`, input.ConversationID, input.SenderID).Scan(&member); err != nil {
		return Message{}, err
	}
	if !member {
		return Message{}, ErrNotMember
	}
	var m Message
	err := s.DB.QueryRow(ctx, `INSERT INTO messages(conversation_id,sender_id,client_message_id,type,text_content) VALUES($1,$2,$3,'text',$4) ON CONFLICT(sender_id,client_message_id) DO UPDATE SET client_message_id=EXCLUDED.client_message_id RETURNING id::text,conversation_id::text,sender_id::text,client_message_id::text,type::text,text_content,created_at`, input.ConversationID, input.SenderID, input.ClientMessageID, strings.TrimSpace(input.Text)).Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.ClientMessageID, &m.Type, &m.Text, &m.CreatedAt)
	if err != nil {
		return Message{}, err
	}
	if err := s.DB.QueryRow(ctx, `SELECT display_name FROM users WHERE id=$1`, m.SenderID).Scan(&m.SenderName); err != nil {
		return Message{}, err
	}
	_, err = s.DB.Exec(ctx, `UPDATE conversations SET last_message_at=$2,updated_at=$2 WHERE id=$1`, m.ConversationID, m.CreatedAt)
	return m, err
}

func (s *Service) List(ctx context.Context, conversationID, userID, before string, limit int) (Page, error) {
	if limit < 1 || limit > 100 {
		limit = 30
	}
	var member bool
	if err := s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM conversation_members WHERE conversation_id=$1 AND user_id=$2)`, conversationID, userID).Scan(&member); err != nil {
		return Page{}, err
	}
	if !member {
		return Page{}, ErrNotMember
	}
	query := `SELECT m.id::text,m.conversation_id::text,m.sender_id::text,u.display_name,m.client_message_id::text,m.type::text,m.text_content,m.created_at,COALESCE(j.id::text,''),COALESCE(j.status::text,''),COALESCE(j.error_message,''),COALESCE(j.title,''),COALESCE(j.duration_seconds,0),COALESCE(j.width,0),COALESCE(j.height,0),COALESCE(j.size_bytes,0) FROM messages m JOIN users u ON u.id=m.sender_id LEFT JOIN media_jobs j ON j.message_id=m.id WHERE m.conversation_id=$1 AND m.deleted_at IS NULL`
	args := []any{conversationID}
	if before != "" {
		if _, err := uuid.Parse(before); err != nil {
			return Page{}, errors.New("invalid cursor")
		}
		cursor := before
		query += ` AND (m.created_at,m.id) < (SELECT created_at,id FROM messages WHERE id=$2 AND conversation_id=$1)`
		args = append(args, cursor)
	}
	query += ` ORDER BY m.created_at DESC,m.id DESC LIMIT $` + itoa(len(args)+1)
	args = append(args, limit+1)
	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return Page{}, err
	}
	defer rows.Close()
	items := make([]Message, 0, limit+1)
	for rows.Next() {
		var m Message
		var jobID, status, jobError, title string
		var duration float64
		var width, height int
		var size int64
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.SenderName, &m.ClientMessageID, &m.Type, &m.Text, &m.CreatedAt, &jobID, &status, &jobError, &title, &duration, &width, &height, &size); err != nil {
			return Page{}, err
		}
		if jobID != "" {
			m.MediaJob = &MediaJob{ID: jobID, Status: status, ErrorMessage: jobError, Title: title, Duration: duration, Width: width, Height: height, Size: size}
			if status == "completed" {
				m.MediaJob.PlaybackURL = "/api/v1/media/jobs/" + jobID + "/content"
			}
		}
		items = append(items, m)
	}

	if err := rows.Err(); err != nil {
		return Page{}, err
	}
	page := Page{HasMore: len(items) > limit}
	if page.HasMore {
		items = items[:limit]
	}
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	if len(items) > 0 && page.HasMore {
		page.NextCursor = items[0].ID
	}
	page.Messages = items
	return page, nil
}

func (s *Service) MarkRead(ctx context.Context, conversationID, userID, messageID string) error {
	cmd, err := s.DB.Exec(ctx, `UPDATE conversation_members cm SET last_read_message_id=m.id,last_read_at=now() FROM messages m WHERE cm.conversation_id=$1 AND cm.user_id=$2 AND m.id=$3 AND m.conversation_id=cm.conversation_id`, conversationID, userID, messageID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func itoa(n int) string { return strconv.Itoa(n) }
