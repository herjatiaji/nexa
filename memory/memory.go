package memory

import (
	"fmt"
	"time"
)

// Fact represents a single remembered user preference or fact.
type Fact struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MemoryStore manages persistent SQLite long-term memory for NEXA.
type MemoryStore struct {
	dbStore *DBStore
}

// NewMemoryStore initializes or loads the memory store using SQLite backend.
func NewMemoryStore() *MemoryStore {
	dbStore, err := NewDBStore()
	if err != nil {
		fmt.Printf("⚠️ SQLite store error: %v (falling back)\n", err)
	}
	return &MemoryStore{
		dbStore: dbStore,
	}
}

// Set saves or updates a fact in SQLite memory.
func (m *MemoryStore) Set(key, value string) error {
	if m.dbStore == nil {
		return fmt.Errorf("database store uninitialized")
	}
	return m.dbStore.Set(key, value)
}

// Get retrieves a specific fact by key.
func (m *MemoryStore) Get(key string) (string, bool) {
	if m.dbStore == nil {
		return "", false
	}
	return m.dbStore.Get(key)
}

// SearchSemantic performs conceptual vector search over stored memories.
func (m *MemoryStore) SearchSemantic(query string) []SemanticMatch {
	if m.dbStore == nil {
		return nil
	}
	facts := m.dbStore.ListFacts()
	return SearchSemantic(facts, query, 0.25)
}

// Delete removes a fact from memory.
func (m *MemoryStore) Delete(key string) error {
	if m.dbStore == nil {
		return fmt.Errorf("database store uninitialized")
	}
	return m.dbStore.Delete(key)
}

// List returns all stored facts as a map.
func (m *MemoryStore) List() map[string]string {
	if m.dbStore == nil {
		return make(map[string]string)
	}
	return m.dbStore.List()
}

// FormatForSystemPrompt formats stored facts for dynamic system prompt injection.
func (m *MemoryStore) FormatForSystemPrompt() string {
	if m.dbStore == nil {
		return "No remembered user facts stored yet."
	}
	return m.dbStore.FormatForSystemPrompt()
}
