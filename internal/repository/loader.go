package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sitanshunandan/glate/internal/domain"
)

// Repository defines how the app accesses substance data.
type Repository interface {
	GetDefinition(id string) (domain.SubstanceDefinition, error)
	GetAll() (map[string]domain.SubstanceDefinition, error)
}

// InMemoryRepo is a thread-safe storage implementation.
type InMemoryRepo struct {
	mu   sync.RWMutex
	data map[string]domain.SubstanceDefinition
}

// NewInMemoryRepo initializes the repo by loading data from a JSON file.
func NewInMemoryRepo(filePath string) (*InMemoryRepo, error) {
	// 1. Open the file
	file, err := os.Open(filePath)
	if err != nil {
		// Try absolute path if relative fails (common issue in Go tests/run)
		absPath, _ := filepath.Abs(filePath)
		return nil, fmt.Errorf("failed to open file at %s: %w", absPath, err)
	}
	defer file.Close()

	// 2. Decode JSON into a temporary slice
	var definitions []domain.SubstanceDefinition
	if err := json.NewDecoder(file).Decode(&definitions); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	// 3. Convert slice to map for O(1) lookups
	dataMap := make(map[string]domain.SubstanceDefinition)
	for _, def := range definitions {
		dataMap[def.ID] = def
	}

	return &InMemoryRepo{
		data: dataMap,
	}, nil
}

// GetDefinition returns a specific substance by ID (Concurrent-safe read).
func (r *InMemoryRepo) GetDefinition(id string) (domain.SubstanceDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	val, ok := r.data[id]
	if !ok {
		return domain.SubstanceDefinition{}, fmt.Errorf("substance '%s' not found", id)
	}
	return val, nil
}

// GetAll returns a copy of the entire map.
func (r *InMemoryRepo) GetAll() (map[string]domain.SubstanceDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification of the internal map
	copyMap := make(map[string]domain.SubstanceDefinition)
	for k, v := range r.data {
		copyMap[k] = v
	}
	return copyMap, nil
}
