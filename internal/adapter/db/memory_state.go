package db

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/upentudu/mockctl/internal/core/ports"
)

const maxDocsPerCollection = 10000

// MemoryStateStore implements ports.StateStore with an ephemeral, thread-safe map.
type MemoryStateStore struct {
	mu    sync.RWMutex
	store map[string]map[string]map[string]any
}

// NewMemoryStateStore initializes an empty state store.
func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{
		store: make(map[string]map[string]map[string]any),
	}
}

func (m *MemoryStateStore) Insert(ctx context.Context, collection string, id string, payload map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.store[collection] == nil {
		m.store[collection] = make(map[string]map[string]any)
	}

	if len(m.store[collection]) >= maxDocsPerCollection {
		return ports.ErrLimitReached
	}

	if id == "" {
		// Attempt to extract from payload, otherwise generate
		if payloadID, ok := payload["id"].(string); ok && payloadID != "" {
			id = payloadID
		} else {
			id = uuid.NewString()
			payload["id"] = id
		}
	}

	payload["createdAt"] = time.Now().UTC().Format(time.RFC3339)
	payload["updatedAt"] = payload["createdAt"]

	m.store[collection][id] = payload
	return nil
}

func (m *MemoryStateStore) Get(ctx context.Context, collection string, id string) (map[string]any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	coll, exists := m.store[collection]
	if !exists {
		return nil, ports.ErrNotFound
	}

	doc, exists := coll[id]
	if !exists {
		return nil, ports.ErrNotFound
	}

	return doc, nil
}

func (m *MemoryStateStore) List(ctx context.Context, collection string) ([]map[string]any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	coll, exists := m.store[collection]
	if !exists {
		return []map[string]any{}, nil
	}

	var results []map[string]any
	for _, doc := range coll {
		results = append(results, doc)
	}

	return results, nil
}

func (m *MemoryStateStore) Update(ctx context.Context, collection string, id string, payload map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	coll, exists := m.store[collection]
	if !exists {
		return ports.ErrNotFound
	}

	if _, exists := coll[id]; !exists {
		return ports.ErrNotFound
	}

	// Preserve createdAt if it exists in the old doc
	if oldDoc, ok := coll[id]; ok {
		if createdAt, ok := oldDoc["createdAt"]; ok {
			payload["createdAt"] = createdAt
		}
	}

	payload["updatedAt"] = time.Now().UTC().Format(time.RFC3339)
	payload["id"] = id

	m.store[collection][id] = payload
	return nil
}

func (m *MemoryStateStore) Patch(ctx context.Context, collection string, id string, partial map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	coll, exists := m.store[collection]
	if !exists {
		return ports.ErrNotFound
	}

	existing, exists := coll[id]
	if !exists {
		return ports.ErrNotFound
	}

	merged := deepMerge(existing, partial)
	merged["updatedAt"] = time.Now().UTC().Format(time.RFC3339)

	m.store[collection][id] = merged
	return nil
}

func (m *MemoryStateStore) Delete(ctx context.Context, collection string, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	coll, exists := m.store[collection]
	if !exists {
		return ports.ErrNotFound
	}

	if _, exists := coll[id]; !exists {
		return ports.ErrNotFound
	}

	delete(coll, id)
	return nil
}

func (m *MemoryStateStore) Reset(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.store = make(map[string]map[string]map[string]any)
	return nil
}

// deepMerge recursively merges src map into dst map.
func deepMerge(dst, src map[string]any) map[string]any {
	out := make(map[string]any)
	for k, v := range dst {
		out[k] = v
	}

	for k, vSrc := range src {
		if vDst, ok := out[k]; ok {
			// If both are maps, recursively merge them
			mapDst, okDst := vDst.(map[string]any)
			mapSrc, okSrc := vSrc.(map[string]any)
			if okDst && okSrc {
				out[k] = deepMerge(mapDst, mapSrc)
				continue
			}
		}
		// Otherwise overwrite with src value
		out[k] = vSrc
	}
	return out
}
