package media

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const TaskDownloadVideo = "media:download-video"

type Job struct {
	ID             string   `json:"id"`
	ConversationID string   `json:"conversation_id"`
	MessageID      string   `json:"message_id,omitempty"`
	RequestedBy    string   `json:"requested_by,omitempty"`
	SourceURL      string   `json:"source_url"`
	Provider       Provider `json:"provider"`
	Status         string   `json:"status,omitempty"`
	Attempts       int      `json:"attempts,omitempty"`
	ErrorCode      string   `json:"error_code,omitempty"`
	ErrorMessage   string   `json:"error_message,omitempty"`
	Title          string   `json:"title,omitempty"`
	Duration       *float64 `json:"duration_seconds,omitempty"`
	Width          *int     `json:"width,omitempty"`
	Height         *int     `json:"height,omitempty"`
	Size           *int64   `json:"size_bytes,omitempty"`
	PlaybackURL    string   `json:"playback_url,omitempty"`
}

type Queue struct{ Client *asynq.Client }

func EnqueueDownload(ctx context.Context, q Queue, job Job) error {
	return enqueueDownload(ctx, q, job, job.ID)
}

func enqueueDownload(ctx context.Context, q Queue, job Job, taskID string) error {
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}
	_, err = q.Client.EnqueueContext(ctx, asynq.NewTask(TaskDownloadVideo, body), asynq.MaxRetry(3), asynq.Timeout(6*time.Minute), asynq.TaskID(taskID))
	return err
}
func (q Queue) Close() error {
	if q.Client == nil {
		return nil
	}
	return q.Client.Close()
}

type JobStore struct{ DB *pgxpool.Pool }

func (s JobStore) CreateForMessage(ctx context.Context, conversationID, messageID, userID, source string, provider Provider) (Job, error) {
	if _, err := url.ParseRequestURI(source); err != nil {
		return Job{}, err
	}
	var j Job
	err := s.DB.QueryRow(ctx, `INSERT INTO media_jobs(conversation_id,message_id,requested_by,source_url,provider) VALUES($1,$2,$3,$4,$5) RETURNING id::text,conversation_id::text,COALESCE(message_id::text,''),requested_by::text,source_url,provider::text`, conversationID, messageID, userID, source, provider).Scan(&j.ID, &j.ConversationID, &j.MessageID, &j.RequestedBy, &j.SourceURL, &j.Provider)
	return j, err
}

func (s JobStore) Create(ctx context.Context, conversationID, userID, clientMessageID, source string, provider Provider) (Job, string, bool, error) {
	var job Job
	var messageID string
	created := false
	err := pgx.BeginTxFunc(ctx, s.DB, pgx.TxOptions{}, func(tx pgx.Tx) error {
		var member bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM conversation_members WHERE conversation_id=$1 AND user_id=$2)`, conversationID, userID).Scan(&member); err != nil {
			return err
		}
		if !member {
			return errors.New("conversation not found")
		}
		err := tx.QueryRow(ctx, `SELECT m.id::text,j.id::text,j.conversation_id::text,j.requested_by::text,j.source_url,j.provider::text,j.status::text,j.attempts,COALESCE(j.error_code,''),COALESCE(j.error_message,'') FROM messages m JOIN media_jobs j ON j.message_id=m.id WHERE m.sender_id=$1 AND m.client_message_id=$2`, userID, clientMessageID).Scan(&messageID, &job.ID, &job.ConversationID, &job.RequestedBy, &job.SourceURL, &job.Provider, &job.Status, &job.Attempts, &job.ErrorCode, &job.ErrorMessage)
		if err == nil {
			if job.ConversationID != conversationID || job.SourceURL != source {
				return errors.New("client_message_id already belongs to another message")
			}
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		var existingConversation, existingType string
		err = tx.QueryRow(ctx, `SELECT id::text,conversation_id::text,type::text FROM messages WHERE sender_id=$1 AND client_message_id=$2`, userID, clientMessageID).Scan(&messageID, &existingConversation, &existingType)
		if err == nil {
			return errors.New("client_message_id already belongs to another message")
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		err = tx.QueryRow(ctx, `INSERT INTO messages(conversation_id,sender_id,client_message_id,type,text_content) VALUES($1,$2,$3,'video',$4) RETURNING id::text`, conversationID, userID, clientMessageID, source).Scan(&messageID)
		if err != nil {
			return err
		}
		err = tx.QueryRow(ctx, `INSERT INTO media_jobs(conversation_id,message_id,requested_by,source_url,provider) VALUES($1,$2,$3,$4,$5) RETURNING id::text,conversation_id::text,message_id::text,requested_by::text,source_url,provider::text,status::text,attempts`, conversationID, messageID, userID, source, provider).Scan(&job.ID, &job.ConversationID, &job.MessageID, &job.RequestedBy, &job.SourceURL, &job.Provider, &job.Status, &job.Attempts)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE conversations SET last_message_at=now(),updated_at=now() WHERE id=$1`, conversationID)
		created = true
		return err
	})
	return job, messageID, created, err
}

func (s JobStore) Complete(ctx context.Context, job Job, key, mime string, size int64, metadata Metadata) error {
	result, err := s.DB.Exec(ctx, `UPDATE media_jobs SET status='completed',storage_key=$2,mime_type=$3,size_bytes=$4,title=$5,duration_seconds=$6,width=$7,height=$8,completed_at=now() WHERE id=$1 AND status='uploading'`, job.ID, key, mime, size, metadata.Title, metadata.Duration, metadata.Width, metadata.Height)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("media job is no longer uploading")
	}
	return nil
}

func (s JobStore) Retry(ctx context.Context, id string) error {
	result, err := s.DB.Exec(ctx, `UPDATE media_jobs SET status='queued',error_code=NULL,error_message=NULL,started_at=NULL WHERE id=$1 AND status='failed' AND attempts < 3`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("media job is not retryable")
	}
	return nil
}

func (s JobStore) SetStatus(ctx context.Context, id, status string, errCode, errMessage string) error {
	var result pgconn.CommandTag
	var err error
	switch status {
	case "failed":
		result, err = s.DB.Exec(ctx, `UPDATE media_jobs SET status=$2,attempts=attempts+1,error_code=$3,error_message=$4 WHERE id=$1 AND status NOT IN ('completed','cancelled','failed')`, id, status, errCode, errMessage)
	case "cancelled":
		result, err = s.DB.Exec(ctx, `UPDATE media_jobs SET status=$2 WHERE id=$1 AND status IN ('queued','fetching_metadata','downloading','processing','uploading')`, id, status)
	default:
		result, err = s.DB.Exec(ctx, `UPDATE media_jobs SET status=$2,started_at=COALESCE(started_at,now()) WHERE id=$1 AND status NOT IN ('completed','cancelled','failed')`, id, status)
	}
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("media job status transition rejected")
	}
	return nil
}
