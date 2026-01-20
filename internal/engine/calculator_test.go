package engine

import (
	"testing"
	"time"
)

func TestRemainingAmount(t *testing.T) {
	calc := NewMetabolicCalculator()

	// Scenario: 100mg of a drug with 1 hour half-life.
	// After 1 hour, exactly 50mg should remain.
	initial := 100.0
	halfLife := 1.0
	elapsed := 1 * time.Hour

	remaining := calc.RemainingAmount(initial, halfLife, elapsed)

	// Floating point comparison needs a small epsilon
	if remaining < 49.9 || remaining > 50.1 {
		t.Errorf("Expected ~50mg, got %f", remaining)
	}
}

func TestTimeUntilClearance(t *testing.T) {
	calc := NewMetabolicCalculator()

	// Scenario: 100mg -> 25mg (2 half-lives).
	// With 2 hour half-life, this should take 4 hours.
	current := 100.0
	target := 25.0
	halfLife := 2.0

	wait := calc.TimeUntilClearance(current, target, halfLife)

	if wait.Hours() < 3.9 || wait.Hours() > 4.1 {
		t.Errorf("Expected ~4 hours, got %f", wait.Hours())
	}
}
