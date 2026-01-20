package engine

import (
	"fmt"
	"time"

	"github.com/sitanshunandan/glate/internal/domain"
	"github.com/sitanshunandan/glate/internal/repository"
)

// Conflict represents a detected issue between two substances.
type Conflict struct {
	SubstanceA string // The active substance (already in body)
	SubstanceB string // The proposed substance
	Type       domain.InteractionType
	Reason     string
	WaitTime   time.Duration // How long until it is safe
}

// Advisor orchestrates the safety checks.
type Advisor struct {
	repo repository.Repository
	calc *MetabolicCalculator
}

// NewAdvisor creates the analysis engine.
func NewAdvisor(repo repository.Repository, calc *MetabolicCalculator) *Advisor {
	return &Advisor{
		repo: repo,
		calc: calc,
	}
}

// CheckSafety evaluates if 'newSubstanceID' can be taken given the 'activeStack'.
func (a *Advisor) CheckSafety(activeStack []domain.ActiveDose, newSubstanceID string) ([]Conflict, error) {
	var conflicts []Conflict
	now := time.Now()

	// 1. Fetch metadata for the proposed substance
	newDef, err := a.repo.GetDefinition(newSubstanceID)
	if err != nil {
		return nil, fmt.Errorf("unknown substance %s: %w", newSubstanceID, err)
	}

	// 2. Iterate through everything currently in the bloodstream
	for _, dose := range activeStack {
		// Fetch metadata for the active dose
		activeDef, err := a.repo.GetDefinition(dose.SubstanceID)
		if err != nil {
			// In production, we might log this and skip, but here we error out
			return nil, fmt.Errorf("unknown active substance %s: %w", dose.SubstanceID, err)
		}

		// Calculate how long it has been in the system
		elapsed := now.Sub(dose.IngestedAt)

		// CHECK A: Does the ACTIVE substance hate the NEW one?
		// e.g., Active Caffeine vs New Iron
		if rule, found := a.findInteraction(activeDef, newSubstanceID); found {
			// Is the window still open?
			window := time.Duration(rule.WindowHours * float64(time.Hour))
			if elapsed < window {
				conflicts = append(conflicts, Conflict{
					SubstanceA: activeDef.Name,
					SubstanceB: newDef.Name,
					Type:       rule.Type,
					Reason:     rule.Note,
					WaitTime:   window - elapsed,
				})
			}
		}

		// CHECK B: Does the NEW substance hate the ACTIVE one?
		// e.g., New DXM vs Active SSRI (Dangerous!)
		if rule, found := a.findInteraction(newDef, dose.SubstanceID); found {
			// For dangerous interactions, we might check if the active dose
			// has effectively cleared (using the Calculator) rather than just a fixed window.
			// For now, we use the fixed window from the new definition.
			window := time.Duration(rule.WindowHours * float64(time.Hour))
			if elapsed < window {
				conflicts = append(conflicts, Conflict{
					SubstanceA: activeDef.Name, // Still list the active one first for clarity
					SubstanceB: newDef.Name,
					Type:       rule.Type,
					Reason:     "Reverse Conflict: " + rule.Note,
					WaitTime:   window - elapsed,
				})
			}
		}
	}

	return conflicts, nil
}

// Helper to search the interaction slice (O(N) is fine here as N is small)
func (a *Advisor) findInteraction(source domain.SubstanceDefinition, targetID string) (domain.Interaction, bool) {
	for _, rule := range source.Interactions {
		if rule.TargetID == targetID {
			return rule, true
		}
	}
	return domain.Interaction{}, false
}
