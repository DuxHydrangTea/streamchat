package realtime

import "testing"

func TestClientJoinRequiresMembership(t *testing.T) {
	hub := NewHub(func(userID, conversationID string) bool { return userID == "u1" && conversationID == "c1" })
	client := &Client{userID: "u2", conversations: map[string]bool{}, hub: hub}
	if client.Join("c1") {
		t.Fatal("non-member joined conversation")
	}
	client.userID = "u1"
	if !client.Join("c1") {
		t.Fatal("member could not join conversation")
	}
	if !client.conversations["c1"] {
		t.Fatal("conversation was not recorded")
	}
}
