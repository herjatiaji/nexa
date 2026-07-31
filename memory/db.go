package memory

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// DBStore manages persistent SQLite database storage for NEXA memories and conversation logs.
type DBStore struct {
	db       *sql.DB
	dbPath   string
	jsonPath string
	mu       sync.RWMutex
}

// NewDBStore initializes SQLite database connection and runs auto-migrations.
func NewDBStore() (*DBStore, error) {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	dbPath := filepath.Join(cwd, "nexa_data.db")
	jsonPath := filepath.Join(cwd, "nexa_memory.json")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database at %s: %w", dbPath, err)
	}

	store := &DBStore{
		db:       db,
		dbPath:   dbPath,
		jsonPath: jsonPath,
	}

	if err := store.initTables(); err != nil {
		return nil, err
	}

	// Auto-migrate old nexa_memory.json if it exists
	_ = store.migrateFromJSON()

	return store, nil
}

func (s *DBStore) initTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS memories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT UNIQUE NOT NULL,
		value TEXT NOT NULL,
		category TEXT DEFAULT 'general',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS conversations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		tool_calls TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS brain_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		type TEXT NOT NULL,
		source TEXT,
		priority INTEGER DEFAULT 5,
		payload TEXT
	);

	CREATE TABLE IF NOT EXISTS cognitive_traces (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cycle_id TEXT NOT NULL,
		stage TEXT NOT NULL,
		component TEXT NOT NULL,
		input TEXT,
		output TEXT,
		duration_ms INTEGER DEFAULT 0,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS brain_snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cycle_id TEXT NOT NULL,
		state TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *DBStore) migrateFromJSON() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.jsonPath)
	if err != nil {
		return nil // No JSON file to migrate
	}

	var factsList []Fact
	if err := json.Unmarshal(data, &factsList); err != nil {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO memories (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at;
	`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, f := range factsList {
		cleanKey := strings.TrimSpace(strings.ToLower(f.Key))
		if cleanKey != "" {
			_, _ = stmt.Exec(cleanKey, f.Value, f.UpdatedAt)
		}
	}

	_ = tx.Commit()
	fmt.Printf("📦 Successfully migrated %d memories from JSON to SQLite database (%s)\n", len(factsList), s.dbPath)
	return nil
}

// Set stores or updates a fact in SQLite.
func (s *DBStore) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleanKey := strings.TrimSpace(strings.ToLower(key))
	query := `
	INSERT INTO memories (key, value, updated_at)
	VALUES (?, ?, ?)
	ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at;
	`
	_, err := s.db.Exec(query, cleanKey, strings.TrimSpace(value), time.Now())
	return err
}

// Get retrieves a fact from SQLite with fuzzy matching.
func (s *DBStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cleanKey := strings.TrimSpace(strings.ToLower(key))
	var val string
	err := s.db.QueryRow("SELECT value FROM memories WHERE key = ?", cleanKey).Scan(&val)
	if err == nil {
		return val, true
	}

	// Substring search
	err = s.db.QueryRow("SELECT value FROM memories WHERE key LIKE ? OR ? LIKE '%' || key || '%'", "%"+cleanKey+"%", cleanKey).Scan(&val)
	if err == nil {
		return val, true
	}

	return "", false
}

// Delete removes a fact from SQLite.
func (s *DBStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleanKey := strings.TrimSpace(strings.ToLower(key))
	_, err := s.db.Exec("DELETE FROM memories WHERE key = ?", cleanKey)
	return err
}

// List returns all stored facts as a map.
func (s *DBStore) List() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT key, value FROM memories ORDER BY key ASC")
	if err != nil {
		return make(map[string]string)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err == nil {
			result[k] = v
		}
	}
	return result
}

// ListFacts returns all stored Fact structs for semantic search.
func (s *DBStore) ListFacts() []Fact {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT key, value, updated_at FROM memories")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var facts []Fact
	for rows.Next() {
		var f Fact
		if err := rows.Scan(&f.Key, &f.Value, &f.UpdatedAt); err == nil {
			facts = append(facts, f)
		}
	}
	return facts
}

// FormatForSystemPrompt formats stored facts for system prompt injection.
func (s *DBStore) FormatForSystemPrompt() string {
	facts := s.ListFacts()
	if len(facts) == 0 {
		return "No remembered user facts stored yet."
	}

	var lines []string
	for _, f := range facts {
		lines = append(lines, fmt.Sprintf("- %s: %s", f.Key, f.Value))
	}
	return strings.Join(lines, "\n")
}

// Close closes the database connection.
func (s *DBStore) Close() error {
	return s.db.Close()
}
