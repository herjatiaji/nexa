package brain

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/heraji/jarvis/autonomy"
	"github.com/heraji/jarvis/cognitive"
	"github.com/heraji/jarvis/goals"
	"github.com/heraji/jarvis/perception"
)

// PersistedSnapshot holds serialized BrainState data for historical replay and debugging.
type PersistedSnapshot struct {
	ID            string                     `json:"id"`
	CycleID       string                     `json:"cycle_id"`
	Context       cognitive.CognitiveContext `json:"context"`
	WorkingMemory cognitive.WorkingMemory    `json:"working_memory"`
	Goals         []goals.Goal               `json:"goals"`
	Social        cognitive.SocialState      `json:"social"`
	Autonomy      autonomy.AutonomyLevel     `json:"autonomy"`
	Percepts      []perception.Percept       `json:"percepts"`
	FusedPercept  *perception.Percept        `json:"fused_percept"`
	PendingIntent *cognitive.CognitiveIntent `json:"pending_intent"`
	CreatedAt     time.Time                  `json:"created_at"`
}

// SnapshotStore keeps an in-memory buffer of key brain snapshots.
type SnapshotStore struct {
	snapshots map[string]PersistedSnapshot
	recentIDs []string
	maxCap    int
	mu        sync.RWMutex
}

func NewSnapshotStore(maxCap int) *SnapshotStore {
	if maxCap <= 0 {
		maxCap = 50
	}
	return &SnapshotStore{
		snapshots: make(map[string]PersistedSnapshot),
		recentIDs: make([]string, 0, maxCap),
		maxCap:    maxCap,
	}
}

// Save creates a PersistedSnapshot from a BrainSnapshot.
func (s *SnapshotStore) Save(cycleID string, bs BrainSnapshot) PersistedSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	ps := PersistedSnapshot{
		ID:            fmt.Sprintf("snap_%s", cycleID),
		CycleID:       cycleID,
		Context:       bs.Context,
		WorkingMemory: bs.WorkingMemory,
		Goals:         bs.Goals,
		Social:        bs.Social,
		Autonomy:      bs.Autonomy,
		Percepts:      bs.Percepts,
		FusedPercept:  bs.FusedPercept,
		PendingIntent: bs.PendingIntent,
		CreatedAt:     time.Now(),
	}

	s.snapshots[cycleID] = ps
	s.recentIDs = append(s.recentIDs, cycleID)

	if len(s.recentIDs) > s.maxCap {
		oldest := s.recentIDs[0]
		delete(s.snapshots, oldest)
		s.recentIDs = s.recentIDs[1:]
	}

	return ps
}

// Get retrieves a snapshot by cycleID.
func (s *SnapshotStore) Get(cycleID string) (PersistedSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ps, ok := s.snapshots[cycleID]
	return ps, ok
}

// ToJSON converts PersistedSnapshot to JSON string.
func (ps PersistedSnapshot) ToJSON() string {
	b, err := json.Marshal(ps)
	if err != nil {
		return "{}"
	}
	return string(b)
}
