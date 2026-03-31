package sync

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"
)

// SyncStatus represents the current state of the mail sync process.
type SyncStatus struct {
	Status    string     `json:"status"`     // "running" | "idle" | "error"
	LastSync  *time.Time `json:"last_sync"`  // nil if never synced
	LastError *string    `json:"last_error"` // nil if no error
}

// SyncManager tracks sync state in memory.
type SyncManager struct {
	mu        sync.Mutex
	status    string
	lastSync  *time.Time
	lastError *string
	syncCmd   string
}

// NewSyncManager creates a new SyncManager with the given sync command.
func NewSyncManager(syncCmd string) *SyncManager {
	return &SyncManager{
		status:  "idle",
		syncCmd: syncCmd,
	}
}

// Status returns the current sync state.
func (s *SyncManager) Status() SyncStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return SyncStatus{
		Status:    s.status,
		LastSync:  s.lastSync,
		LastError: s.lastError,
	}
}

// TriggerSync starts a sync in a background goroutine. Returns an error if already running.
func (s *SyncManager) TriggerSync() error {
	s.mu.Lock()
	if s.status == "running" {
		s.mu.Unlock()
		return fmt.Errorf("sync already running")
	}
	s.status = "running"
	s.lastError = nil
	s.mu.Unlock()

	go s.runSync()
	return nil
}

func (s *SyncManager) runSync() {
	log.Printf("Starting mail sync: %s", s.syncCmd)

	cmd := exec.Command("sh", "-c", s.syncCmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	s.lastSync = &now

	if err != nil {
		errMsg := fmt.Sprintf("%v: %s", err, stderr.String())
		s.lastError = &errMsg
		s.status = "error"
		log.Printf("Mail sync failed: %s", errMsg)
	} else {
		s.lastError = nil
		s.status = "idle"
		log.Printf("Mail sync completed successfully")
	}

	if stdout.Len() > 0 {
		log.Printf("Sync stdout: %s", stdout.String())
	}
	if stderr.Len() > 0 {
		log.Printf("Sync stderr: %s", stderr.String())
	}
}
