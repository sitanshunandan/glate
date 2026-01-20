package domain

import "time"

// InteractionType defines the nature of the relationship between two compounds.
// We use string types for better readability in JSON/DBs.
type InteractionType string

const (
	// Inhibit: Substance A blocks or reduces the absorption of B (e.g., Calcium -> Iron).
	TypeInhibit InteractionType = "INHIBIT"

	// Potentiate: Substance A increases the effect or absorption of B (e.g., Vit C -> Iron).
	TypePotentiate InteractionType = "POTENTIATE"

	// Dangerous: The combination poses a health risk (e.g., MAOIs + SSRIs).
	TypeDangerous InteractionType = "DANGEROUS"
)

// SubstanceCategory helps UI/Logic group items (e.g., "Don't take stimulants after 4 PM").
type SubstanceCategory string

const (
	CatMineral   SubstanceCategory = "Mineral"
	CatVitamin   SubstanceCategory = "Vitamin"
	CatStimulant SubstanceCategory = "Stimulant"
	CatNootropic SubstanceCategory = "Nootropic"
	CatAminoAcid SubstanceCategory = "AminoAcid"
)

// -------------------------------------------------------------------------
// Static Definitions (The "Public Database" Data)
// -------------------------------------------------------------------------

// Interaction represents a rule: "If you take X, be careful with TargetID".
type Interaction struct {
	TargetID    string          `json:"target_id"`    // The ID of the *other* substance
	Type        InteractionType `json:"type"`         // INHIBIT, POTENTIATE, DANGEROUS
	WindowHours float64         `json:"window_hours"` // How long the interaction lasts (clearance window)
	Note        string          `json:"note"`         // Clinical explanation (e.g., "Competes for DMT1 transporter")
}

// SubstanceDefinition is the immutable science data.
// It comes from your JSON seeder or DB.
type SubstanceDefinition struct {
	ID              string            `json:"id"`              // Unique slug (e.g., "magnesium-glycinate")
	Name            string            `json:"name"`            // Display name
	Category        SubstanceCategory `json:"category"`        // e.g., Mineral
	HalfLifeHours   float64           `json:"half_life_hours"` // e.g., 4.0
	Bioavailability float64           `json:"bioavailability"` // 0.0 to 1.0 (Absorption efficiency)
	Interactions    []Interaction     `json:"interactions"`    // The graph edges (dependencies)
}

// -------------------------------------------------------------------------
// Runtime State (The User's Data)
// -------------------------------------------------------------------------

// ActiveDose represents a pill the user has actually swallowed.
type ActiveDose struct {
	ID          string    // UUID for this specific event
	SubstanceID string    // Links back to SubstanceDefinition.ID
	AmountMg    float64   // How much was taken
	IngestedAt  time.Time // Timestamp of ingestion
}
