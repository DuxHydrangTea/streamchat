package conversations

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Conversation struct {
	ID            string   `json:"id"`
	Type          string   `json:"type"`
	Name          string   `json:"name,omitempty"`
	LastMessageAt *string  `json:"last_message_at,omitempty"`
	Members       []Member `json:"members,omitempty"`
}
type Member struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}
type Service struct{ DB *pgxpool.Pool }

func ValidateCreate(kind, name string, members []string) error {
	if kind != "direct" && kind != "group" {
		return errors.New("type must be direct or group")
	}
	if kind == "group" && strings.TrimSpace(name) == "" {
		return errors.New("group name is required")
	}
	if len(name) > 100 {
		return errors.New("name is too long")
	}
	if kind == "direct" && len(members) != 1 {
		return errors.New("direct conversation requires exactly one other member")
	}
	if len(members) > 99 {
		return errors.New("too many members")
	}
	return nil
}

func (s *Service) Create(ctx context.Context, creator, kind, name string, members []string) (Conversation, error) {
	if err := ValidateCreate(kind, name, members); err != nil {
		return Conversation{}, err
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return Conversation{}, err
	}
	defer tx.Rollback(ctx)
	var c Conversation
	err = tx.QueryRow(ctx, `INSERT INTO conversations(type,name,created_by) VALUES($1,$2,$3) RETURNING id::text,type::text,COALESCE(name,'')`, kind, nullIfEmpty(name), creator).Scan(&c.ID, &c.Type, &c.Name)
	if err != nil {
		return Conversation{}, err
	}
	seen := map[string]bool{creator: true}
	if _, err = tx.Exec(ctx, `INSERT INTO conversation_members(conversation_id,user_id,role) VALUES($1,$2,'admin')`, c.ID, creator); err != nil {
		return Conversation{}, err
	}
	for _, id := range members {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		cmd, e := tx.Exec(ctx, `INSERT INTO conversation_members(conversation_id,user_id) SELECT $1,id FROM users WHERE id=$2 ON CONFLICT DO NOTHING`, c.ID, id)
		if e != nil {
			return Conversation{}, e
		}
		if cmd.RowsAffected() == 0 {
			return Conversation{}, errors.New("member not found")
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return Conversation{}, err
	}
	c.Type = kind
	c.Name = name
	return c, nil
}
func (s *Service) List(ctx context.Context, userID string) ([]Conversation, error) {
	rows, err := s.DB.Query(ctx, `SELECT c.id::text,c.type::text,COALESCE(c.name,''),c.last_message_at::text FROM conversations c JOIN conversation_members cm ON cm.conversation_id=c.id WHERE cm.user_id=$1 ORDER BY c.last_message_at DESC NULLS LAST,c.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Conversation{}
	for rows.Next() {
		var c Conversation
		if err := rows.Scan(&c.ID, &c.Type, &c.Name, &c.LastMessageAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
func (s *Service) Members(ctx context.Context, userID, conversationID string) ([]Member, error) {
	var allowed bool
	if err := s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM conversation_members WHERE conversation_id=$1 AND user_id=$2)`, conversationID, userID).Scan(&allowed); err != nil {
		return nil, err
	}
	if !allowed {
		return nil, pgx.ErrNoRows
	}
	rows, err := s.DB.Query(ctx, `SELECT u.id::text,u.username,u.display_name FROM users u JOIN conversation_members cm ON cm.user_id=u.id WHERE cm.conversation_id=$1 ORDER BY cm.joined_at`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Member{}
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Username, &m.DisplayName); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return strings.TrimSpace(s)
}
