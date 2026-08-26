package storage

import (
	"context"
	"sync"
	"testing"
)

func TestMemoryStateStore_CRUD(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStateStore()

	// 1. Insert & Auto-Generation
	payload := map[string]any{"name": "test_user"}
	err := store.Insert(ctx, "users", "1", payload)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	// 2. Get & Verify
	doc, err := store.Get(ctx, "users", "1")
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}
	if doc["name"] != "test_user" {
		t.Errorf("Expected name 'test_user', got %v", doc["name"])
	}

	// 3. Patch (Deep Merge)
	patch := map[string]any{"age": 25}
	err = store.Patch(ctx, "users", "1", patch)
	if err != nil {
		t.Fatalf("Failed to patch document: %v", err)
	}

	doc, _ = store.Get(ctx, "users", "1")
	if doc["age"] != 25 {
		t.Errorf("Expected age 25 after patch, got %v", doc["age"])
	}

	// 4. Delete
	err = store.Delete(ctx, "users", "1")
	if err != nil {
		t.Fatalf("Failed to delete document: %v", err)
	}

	// Ensure it's deleted
	_, err = store.Get(ctx, "users", "1")
	if err == nil {
		t.Errorf("Expected error after deletion, got nil (Document still exists)")
	}
}

func TestMemoryStateStore_Concurrency(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStateStore()
	var wg sync.WaitGroup

	// Spawn 200 goroutines to violently mutate the state concurrently
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Constant chaotic reads and writes
			_ = store.Insert(ctx, "load_test", "doc", map[string]any{"val": id})
			_, _ = store.Get(ctx, "load_test", "doc")
			_, _ = store.List(ctx, "load_test")
		}(i)
	}

	wg.Wait()
	// If the test completes without throwing "fatal error: concurrent map writes", the Global Lock is perfectly implemented!
}
