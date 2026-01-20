package store

import (
	"sync"

	"github.com/sitanshunandan/glate/internal/domain"
)

// SessionStore manages user state in memory.
// In a distributed system, this would be Redis.
type SessionStore struct {
	mu sync.RWMutex
	// Map of UserID -> Slice of ActiveDoses
	stacks map[string][]domain.ActiveDose
}

// NewSessionStore initializes the storage.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		stacks: make(map[string][]domain.ActiveDose),
	}
}

// AddDose safely appends a dose to the user's stack.
func (s *SessionStore) AddDose(userID string, dose domain.ActiveDose) {
	s.mu.Lock()         // 🔒 LOCK for writing
	defer s.mu.Unlock() // 🔓 UNLOCK when done

	s.stacks[userID] = append(s.stacks[userID], dose)
}

// GetStack safely retrieves a copy of the user's stack.
func (s *SessionStore) GetStack(userID string) []domain.ActiveDose {
	s.mu.RLock()         // 🔒 READ LOCK (allows other readers)
	defer s.mu.RUnlock() // 🔓 UNLOCK

	// Return a copy to prevent race conditions if the caller modifies it
	original := s.stacks[userID]
	copyStack := make([]domain.ActiveDose, len(original))
	copy(copyStack, original)

	return copyStack
}

// ClearStack resets the user (e.g., for a new day).
func (s *SessionStore) ClearStack(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.stacks, userID)
}
