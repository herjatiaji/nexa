package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Fact represents a single remembered user preference or fact.
type Fact struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MemoryStore manages persistent long-term memory for NEXA.
type MemoryStore struct {
	filePath string
	facts    map[string]Fact
	mu       sync.RWMutex
}

// NewMemoryStore initializes or loads the memory store from disk.
func NewMemoryStore() *MemoryStore {
	memPath := "nexa_memory.json"
	if cwd, err := os.Getwd(); err == nil {
		memPath = filepath.Join(cwd, "nexa_memory.json")
	}

	store := &MemoryStore{
		filePath: memPath,
		facts:    make(map[string]Fact),
	}
	_ = store.load()
	return store
}

func (m *MemoryStore) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return err
	}

	var factsList []Fact
	if err := json.Unmarshal(data, &factsList); err != nil {
		return err
	}

	for _, f := range factsList {
		m.facts[strings.ToLower(f.Key)] = f
	}
	return nil
}

func (m *MemoryStore) saveLocked() error {
	var factsList []Fact
	for _, f := range m.facts {
		factsList = append(factsList, f)
	}

	data, err := json.MarshalIndent(factsList, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0644)
}

// Set saves or updates a fact in memory.
func (m *MemoryStore) Set(key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cleanKey := strings.TrimSpace(strings.ToLower(key))
	m.facts[cleanKey] = Fact{
		Key:       cleanKey,
		Value:     strings.TrimSpace(value),
		UpdatedAt: time.Now(),
	}
	return m.saveLocked()
}

// Get retrieves a specific fact by key (supports fuzzy/normalized matching).
func (m *MemoryStore) Get(key string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cleanKey := strings.TrimSpace(strings.ToLower(key))
	if f, ok := m.facts[cleanKey]; ok {
		return f.Value, true
	}

	// Try normalized underscore/space replacement
	altKey := strings.ReplaceAll(cleanKey, "_", " ")
	if f, ok := m.facts[altKey]; ok {
		return f.Value, true
	}

	altKey2 := strings.ReplaceAll(cleanKey, " ", "_")
	if f, ok := m.facts[altKey2]; ok {
		return f.Value, true
	}

	// Substring fallback
	for k, f := range m.facts {
		if strings.Contains(k, cleanKey) || strings.Contains(cleanKey, k) {
			return f.Value, true
		}
	}

	return "", false
}

// Delete removes a fact from memory.
func (m *MemoryStore) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cleanKey := strings.TrimSpace(strings.ToLower(key))
	delete(m.facts, cleanKey)
	return m.saveLocked()
}

// List returns all stored facts as a map.
func (m *MemoryStore) List() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range m.facts {
		result[k] = v.Value
	}
	return result
}

// FormatForSystemPrompt formats stored facts for dynamic system prompt injection.
func (m *MemoryStore) FormatForSystemPrompt() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.facts) == 0 {
		return "No remembered user facts stored yet."
	}

	var lines []string
	for _, f := range m.facts {
		lines = append(lines, fmt.Sprintf("- %s: %s", f.Key, f.Value))
	}
	return strings.Join(lines, "\n")
}
