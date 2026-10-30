package realtime

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/gorilla/websocket"
)

const queueSize = 64
const maxConnectionsPerUser = 5

type Event struct {
	EventID        string    `json:"event_id"`
	Type           string    `json:"event_type"`
	ConversationID string    `json:"conversation_id"`
	Payload        any       `json:"payload"`
	Timestamp      time.Time `json:"timestamp"`
}
type outbound struct{ event Event }
type Client struct {
	removed       atomic.Bool
	conn          *websocket.Conn
	userID        string
	conversations map[string]bool
	send          chan outbound
	hub           *Hub
}
type Hub struct {
	mu         sync.RWMutex
	rooms      map[string]map[*Client]bool
	users      map[string]int
	membership func(userID, conversationID string) bool
}

func NewHub(membership func(string, string) bool) *Hub {
	return &Hub{rooms: map[string]map[*Client]bool{}, users: map[string]int{}, membership: membership}
}
func (h *Hub) Add(userID string, conn *websocket.Conn) (*Client, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.users[userID] >= maxConnectionsPerUser {
		return nil, false
	}
	h.users[userID]++
	return &Client{conn: conn, userID: userID, conversations: map[string]bool{}, send: make(chan outbound, queueSize), hub: h}, true
}
func (h *Hub) Remove(c *Client) {
	if !c.removed.CompareAndSwap(false, true) {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for room := range c.conversations {
		delete(h.rooms[room], c)
		if len(h.rooms[room]) == 0 {
			delete(h.rooms, room)
		}
	}
	if h.users[c.userID] > 1 {
		h.users[c.userID]--
	} else {
		delete(h.users, c.userID)
	}
	close(c.send)
	_ = c.conn.Close()
}
func (c *Client) Join(conversationID string) bool {
	if !c.hub.membership(c.userID, conversationID) {
		return false
	}
	c.hub.mu.Lock()
	defer c.hub.mu.Unlock()
	c.conversations[conversationID] = true
	if c.hub.rooms[conversationID] == nil {
		c.hub.rooms[conversationID] = map[*Client]bool{}
	}
	c.hub.rooms[conversationID][c] = true
	return true
}
func (h *Hub) Publish(conversationID string, event Event) {
	if event.EventID == "" {
		event.EventID = uuid.NewString()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.rooms[conversationID] {
		select {
		case c.send <- outbound{event: event}:
		default:
			go c.hub.Remove(c)
		}
	}
}
func (c *Client) WritePump(done chan struct{}) {
	defer c.hub.Remove(c)
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	defer close(done)
	for {
		select {
		case item, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			data, err := json.Marshal(item.event)
			if err != nil {
				return
			}
			if err = c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.WriteMessage(websocket.PingMessage, nil)
		}
	}
}
func (c *Client) ReadPump() {
	defer c.hub.Remove(c)
	c.conn.SetReadLimit(4096)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error { return c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)) })
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		var msg struct {
			Action         string `json:"action"`
			ConversationID string `json:"conversation_id"`
		}
		if json.Unmarshal(data, &msg) != nil {
			continue
		}
		if msg.Action == "join" {
			c.Join(msg.ConversationID)
		}
	}
}
