package engine

import (
	"math"
	"time"
)

// MetabolicCalculator handles the pharmacokinetic math.
// It is stateless and thread-safe.
type MetabolicCalculator struct{}

// NewMetabolicCalculator creates a new instance.
func NewMetabolicCalculator() *MetabolicCalculator {
	return &MetabolicCalculator{}
}

// -------------------------------------------------------------------------
// Core Math: First-Order Kinetics
// Formula: Ct = C0 * e^(-kt)
// -------------------------------------------------------------------------

// RemainingAmount calculates how much of a substance remains active.
// initialMg: The dose taken (e.g., 500mg)
// halfLifeHours: The substance's biological half-life
// elapsed: How much time has passed since ingestion
func (c *MetabolicCalculator) RemainingAmount(initialMg float64, halfLifeHours float64, elapsed time.Duration) float64 {
	// 1. Edge Case: If time is negative (future dose) or zero, return initial
	if elapsed <= 0 {
		return initialMg
	}

	// 2. Calculate Elimination Rate Constant (k)
	// k = ln(2) / t_1/2
	k := math.Log(2) / halfLifeHours

	// 3. Apply Decay Formula
	// We convert duration to float hours to match the half-life unit
	elapsedHours := elapsed.Hours()
	remaining := initialMg * math.Exp(-k*elapsedHours)

	return remaining
}

// TimeUntilClearance calculates how long until the substance drops below a specific threshold.
// helpful for: "When will my Caffeine be low enough to sleep?"
func (c *MetabolicCalculator) TimeUntilClearance(currentMg float64, targetMg float64, halfLifeHours float64) time.Duration {
	if currentMg <= targetMg {
		return 0
	}

	// Rearranging the decay formula to solve for t:
	// t = -ln(Ct / C0) / k
	k := math.Log(2) / halfLifeHours

	hoursNeeded := -math.Log(targetMg/currentMg) / k

	return time.Duration(hoursNeeded * float64(time.Hour))
}
