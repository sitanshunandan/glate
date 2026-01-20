package engine

import (
	"fmt"
	"log"
	"time"

	"github.com/sitanshunandan/glate/internal/repository"
	"github.com/sitanshunandan/glate/internal/store"
)

// Monitor runs background checks on active sessions.
type Monitor struct {
	Store *store.SessionStore
	Repo  repository.Repository
	Calc  *MetabolicCalculator
}

// NewMonitor creates the background worker.
func NewMonitor(store *store.SessionStore, repo repository.Repository, calc *MetabolicCalculator) *Monitor {
	return &Monitor{Store: store, Repo: repo, Calc: calc}
}

// Start begins the monitoring loop in a non-blocking Goroutine.
func (m *Monitor) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)

	// Launch the worker in the background (The "Golang Flex")
	go func() {
		log.Printf("👁️  Background Monitor online. Scanning every %s...", interval)
		for range ticker.C {
			m.scan()
		}
	}()
}

// scan iterates over all users safely and logs their status.
func (m *Monitor) scan() {
	sessions := m.Store.GetAllSessions()
	now := time.Now()

	// If no one is active, keep the logs quiet
	if len(sessions) == 0 {
		return
	}

	fmt.Println("\n--- 🏥 System Heartbeat ---")
	for userID, stack := range sessions {
		if len(stack) == 0 {
			continue
		}
		fmt.Printf("User [%s]:\n", userID)

		for _, dose := range stack {
			def, err := m.Repo.GetDefinition(dose.SubstanceID)
			if err != nil {
				continue
			}

			elapsed := now.Sub(dose.IngestedAt)
			remaining := m.Calc.RemainingAmount(dose.AmountMg, def.HalfLifeHours, elapsed)

			// Fancy formatting: Visual bar for decay
			// If remaining > 50%, show green. If low, show yellow.
			fmt.Printf("   • %-20s | Original: %.0fmg | Current: %.1fmg (T+%.0fm)\n",
				def.Name, dose.AmountMg, remaining, elapsed.Minutes())

			// ALERT LOGIC:
			// If a stimulant drops below 50mg, log a "Sleep Window" alert
			if def.Category == "Stimulant" && remaining < 50.0 && remaining > 40.0 {
				fmt.Printf("     💤 SLEEP WINDOW OPEN: %s is low enough.\n", def.Name)
			}
		}
	}
	fmt.Println("---------------------------")
}
