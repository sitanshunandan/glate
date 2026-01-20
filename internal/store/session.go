package store

import (
	"sync"

	"github.com/sitanshunandan/glate/internal/domain"
)

// SessionStore manages user state in memory.
type SessionStore struct {
	mu     sync.RWMutex
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
	s.mu.Lock()
	defer s.mu.Unlock()

	// Initialize slice if it doesn't exist
	if _, ok := s.stacks[userID]; !ok {
		s.stacks[userID] = []domain.ActiveDose{}
	}

	s.stacks[userID] = append(s.stacks[userID], dose)
}

// GetStack safely retrieves a copy of the user's stack.
func (s *SessionStore) GetStack(userID string) []domain.ActiveDose {
	s.mu.RLock()
	defer s.mu.RUnlock()

	original, ok := s.stacks[userID]
	if !ok {
		return []domain.ActiveDose{}
	}

	// Return a deep copy to prevent race conditions
	copyStack := make([]domain.ActiveDose, len(original))
	copy(copyStack, original)

	return copyStack
}

// GetAllSessions returns a snapshot of all active users.
// This is the method that was missing!
func (s *SessionStore) GetAllSessions() map[string][]domain.ActiveDose {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot := make(map[string][]domain.ActiveDose)
	for userID, stack := range s.stacks {
		// Deep copy the slice
		userStack := make([]domain.ActiveDose, len(stack))
		copy(userStack, stack)
		snapshot[userID] = userStack
	}

	return snapshot
}

// ClearStack resets the user (optional utility).
func (s *SessionStore) ClearStack(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.stacks, userID)
}
