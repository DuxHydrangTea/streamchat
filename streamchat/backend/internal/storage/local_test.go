package storage

import (
	"context"
	"strings"
	"testing"
)

func TestLocalStoreRoundTrip(t *testing.T) {
	store := Local{Root: t.TempDir()}
	if err := store.Put(context.Background(), "uploads/a.txt", strings.NewReader("hello"), "text/plain", 5); err != nil {
		t.Fatal(err)
	}
	f, err := store.Open(context.Background(), "uploads/a.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	buf := make([]byte, 5)
	if _, err := f.Read(buf); err != nil {
		t.Fatal(err)
	}
	if string(buf) != "hello" {
		t.Fatalf("got %q", buf)
	}
	if err := store.Delete(context.Background(), "uploads/a.txt"); err != nil {
		t.Fatal(err)
	}
}
func TestLocalStoreRejectsTraversal(t *testing.T) {
	store := Local{Root: t.TempDir()}
	if err := store.Put(context.Background(), "../escape", strings.NewReader("x"), "text/plain", 1); err == nil {
		t.Fatal("expected traversal to be rejected")
	}
}
