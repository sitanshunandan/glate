package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sitanshunandan/glate/internal/domain"
	"github.com/sitanshunandan/glate/internal/engine"
	"github.com/sitanshunandan/glate/internal/repository"
)

func main() {
	fmt.Println("--- Glate: Metabolic Engine Initializing ---")

	// 1. Setup Data Layer
	configPath := "configs/substances.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file missing: %s", configPath)
	}

	repo, err := repository.NewInMemoryRepo(configPath)
	if err != nil {
		log.Fatalf("Repo failure: %v", err)
	}

	// 2. Setup Logic Layer
	calc := engine.NewMetabolicCalculator()
	advisor := engine.NewAdvisor(repo, calc)

	// ---------------------------------------------------------
	// SCENARIO 1: The "Morning Coffee" Problem
	// ---------------------------------------------------------
	fmt.Println("\n--- Scenario: Morning Routine ---")

	// User took Caffeine 30 minutes ago
	activeStack := []domain.ActiveDose{
		{
			ID:          "dose-1",
			SubstanceID: "caffeine",
			AmountMg:    150,
			IngestedAt:  time.Now().Add(-30 * time.Minute),
		},
	}
	fmt.Println("inputs: User took 150mg Caffeine 30 mins ago.")

	// User wants to take Iron Bisglycinate NOW
	proposed := "iron-bisglycinate"
	fmt.Printf("Action: User wants to take %s.\n", proposed)

	// Ask the Advisor
	conflicts, err := advisor.CheckSafety(activeStack, proposed)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	// Output Results
	if len(conflicts) > 0 {
		fmt.Println("\n❌ BLOCKED! Conflicts detected:")
		for _, c := range conflicts {
			fmt.Printf("   [%s] %s -> %s\n", c.Type, c.SubstanceA, c.SubstanceB)
			fmt.Printf("   Reason: %s\n", c.Reason)
			// Round duration for cleaner output
			wait := c.WaitTime.Round(time.Minute)
			fmt.Printf("   ⚠️  Please wait %s before taking.\n", wait)
		}
	} else {
		fmt.Println("\n✅ SAFE. No interactions detected.")
	}
}
