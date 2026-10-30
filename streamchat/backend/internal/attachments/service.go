package attachments

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"streamchat/backend/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

const MaxUploadBytes int64 = 200 * 1024 * 1024

var allowedMIME = map[string]bool{"image/jpeg": true, "image/png": true, "image/gif": true, "video/mp4": true, "application/pdf": true, "text/plain": true}

type Attachment struct {
	ID         string `json:"id"`
	StorageKey string `json:"storage_key"`
	Filename   string `json:"original_filename"`
	MimeType   string `json:"mime_type"`
	Size       int64  `json:"size_bytes"`
	Status     string `json:"status"`
}
type Service struct {
	DB    *pgxpool.Pool
	Store storage.Store
	Root  string
}

func ValidateUpload(filename, mime string, size int64) error {
	if strings.TrimSpace(filename) == "" || len(filename) > 255 {
		return errors.New("invalid filename")
	}
	if filepath.Base(filename) != filename {
		return errors.New("invalid filename")
	}
	if !allowedMIME[mime] {
		return errors.New("file type is not allowed")
	}
	if size < 1 || size > MaxUploadBytes {
		return errors.New("file size is not allowed")
	}
	return nil
}
func newKey(userID string) string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("uploads/%s/%s", userID, hex.EncodeToString(b[:]))
}
func (s *Service) Create(ctx context.Context, userID, conversationID, filename, mime string, size int64) (Attachment, error) {
	if err := ValidateUpload(filename, mime, size); err != nil {
		return Attachment{}, err
	}
	var member bool
	if err := s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM conversation_members WHERE conversation_id=$1 AND user_id=$2)`, conversationID, userID).Scan(&member); err != nil {
		return Attachment{}, err
	}
	if !member {
		return Attachment{}, errors.New("conversation not found")
	}
	key := newKey(userID)
	var a Attachment
	err := s.DB.QueryRow(ctx, `INSERT INTO attachments(conversation_id,uploaded_by,storage_key,original_filename,mime_type,size_bytes) VALUES($1,$2,$3,$4,$5,$6) RETURNING id::text,storage_key,original_filename,mime_type,size_bytes,status::text`, conversationID, userID, key, filename, mime, size).Scan(&a.ID, &a.StorageKey, &a.Filename, &a.MimeType, &a.Size, &a.Status)
	return a, err
}
func (s *Service) Complete(ctx context.Context, userID, id string, size int64) (Attachment, error) {
	var a Attachment
	if err := s.DB.QueryRow(ctx, `SELECT id::text,storage_key,original_filename,mime_type,size_bytes FROM attachments WHERE id=$1 AND uploaded_by=$2 AND status='pending'`, id, userID).Scan(&a.ID, &a.StorageKey, &a.Filename, &a.MimeType, &a.Size); err != nil {
		return Attachment{}, err
	}
	if size != a.Size {
		return Attachment{}, errors.New("uploaded size does not match declared size")
	}
	object, err := s.Store.Open(ctx, a.StorageKey)
	if err != nil {
		return Attachment{}, err
	}
	actual, err := io.Copy(io.Discard, object)
	_ = object.Close()
	if err != nil {
		return Attachment{}, err
	}
	if actual != a.Size {
		return Attachment{}, errors.New("uploaded object size does not match")
	}
	err = s.DB.QueryRow(ctx, `UPDATE attachments SET status='completed',completed_at=now() WHERE id=$1 AND uploaded_by=$2 AND status='pending' RETURNING id::text,storage_key,original_filename,mime_type,size_bytes,status::text`, id, userID).Scan(&a.ID, &a.StorageKey, &a.Filename, &a.MimeType, &a.Size, &a.Status)
	return a, err
}
func (s *Service) Abort(ctx context.Context, userID, id string) error {
	_, err := s.DB.Exec(ctx, `UPDATE attachments SET status='aborted' WHERE id=$1 AND uploaded_by=$2 AND status='pending'`, id, userID)
	return err
}
func (s *Service) Access(ctx context.Context, userID, id string) (Attachment, error) {
	var a Attachment
	err := s.DB.QueryRow(ctx, `SELECT a.id::text,a.storage_key,a.original_filename,a.mime_type,a.size_bytes,a.status::text FROM attachments a JOIN conversation_members cm ON cm.conversation_id=a.conversation_id WHERE a.id=$1 AND cm.user_id=$2 AND a.status='completed'`, id, userID).Scan(&a.ID, &a.StorageKey, &a.Filename, &a.MimeType, &a.Size, &a.Status)
	return a, err
}
func (s *Service) Put(ctx context.Context, a Attachment, r io.Reader) error {
	limited := &io.LimitedReader{R: r, N: MaxUploadBytes + 1}
	if err := s.Store.Put(ctx, a.StorageKey, limited, a.MimeType, a.Size); err != nil {
		return err
	}
	if limited.N <= 0 {
		_ = s.Store.Delete(ctx, a.StorageKey)
		return errors.New("file exceeds size limit")
	}
	return nil
}
