package security

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// PermissionLevel defines access rule (ALLOW, CONFIRM, DENY).
type PermissionLevel string

const (
	LevelAllow   PermissionLevel = "ALLOW"
	LevelConfirm PermissionLevel = "CONFIRM"
	LevelDeny    PermissionLevel = "DENY"
)

// PermissionManager controls OS-level application capabilities for NEXA.
type PermissionManager struct {
	permissions map[string]PermissionLevel
	filePath    string
	mu          sync.RWMutex
}

// NewPermissionManager creates and initializes permission rules from disk.
func NewPermissionManager() *PermissionManager {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	permPath := filepath.Join(cwd, "nexa_permissions.json")

	pm := &PermissionManager{
		permissions: map[string]PermissionLevel{
			"desktop_apps":     LevelAllow,
			"filesystem.read":  LevelAllow,
			"filesystem.write": LevelAllow,
			"filesystem.delete": LevelConfirm,
			"terminal.execute": LevelConfirm,
			"vision":           LevelAllow,
			"web":              LevelAllow,
			"mcp":              LevelAllow,
			"memory":           LevelAllow,
		},
		filePath: permPath,
	}

	_ = pm.load()
	return pm
}

func (pm *PermissionManager) load() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	data, err := os.ReadFile(pm.filePath)
	if err != nil {
		return nil
	}

	var loaded map[string]PermissionLevel
	if err := json.Unmarshal(data, &loaded); err == nil {
		for k, v := range loaded {
			pm.permissions[k] = v
		}
	}
	return nil
}

func (pm *PermissionManager) saveLocked() error {
	data, err := json.MarshalIndent(pm.permissions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(pm.filePath, data, 0644)
}

// CheckPermission evaluates access level for a target capability.
func (pm *PermissionManager) CheckPermission(capability string) PermissionLevel {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	cleanCap := strings.ToLower(strings.TrimSpace(capability))
	if lvl, ok := pm.permissions[cleanCap]; ok {
		return lvl
	}
	return LevelAllow
}

// SetPermission updates access level for a capability and persists it.
func (pm *PermissionManager) SetPermission(capability string, level PermissionLevel) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	cleanCap := strings.ToLower(strings.TrimSpace(capability))
	pm.permissions[cleanCap] = level
	return pm.saveLocked()
}

// ListPermissions returns all permission settings.
func (pm *PermissionManager) ListPermissions() map[string]PermissionLevel {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]PermissionLevel)
	for k, v := range pm.permissions {
		result[k] = v
	}
	return result
}
