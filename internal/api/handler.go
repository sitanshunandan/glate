package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sitanshunandan/glate/internal/domain"
	"github.com/sitanshunandan/glate/internal/engine"
	"github.com/sitanshunandan/glate/internal/repository"
	"github.com/sitanshunandan/glate/internal/store"
)

// -------------------------------------------------------------------------
// Request DTOs
// -------------------------------------------------------------------------

// AnalysisRequest is for the stateless "Check Safety" endpoint.
type AnalysisRequest struct {
	ActiveStack []ActiveDoseDTO `json:"active_stack"`
	ProposedID  string          `json:"proposed_id"`
}

// ActiveDoseDTO helps us parse JSON time strings safely.
type ActiveDoseDTO struct {
	SubstanceID   string  `json:"substance_id"`
	AmountMg      float64 `json:"amount_mg"`
	IngestedAtStr string  `json:"ingested_at"`
}

// IngestRequest is for the stateful "Take Pill" endpoint.
type IngestRequest struct {
	UserID      string  `json:"user_id"`
	SubstanceID string  `json:"substance_id"`
	AmountMg    float64 `json:"amount_mg"`
}

// -------------------------------------------------------------------------
// Handler & Factory
// -------------------------------------------------------------------------

// Handler holds the dependencies.
type Handler struct {
	Advisor *engine.Advisor
	Store   *store.SessionStore
	Repo    repository.Repository       // <--- NEW
	Calc    *engine.MetabolicCalculator // <--- NEW
}

// NewHandler injects dependencies.
func NewHandler(advisor *engine.Advisor, store *store.SessionStore, repo repository.Repository, calc *engine.MetabolicCalculator) *Handler {
	return &Handler{
		Advisor: advisor,
		Store:   store,
		Repo:    repo,
		Calc:    calc,
	}
}

type StatusResponse struct {
	Substance   string  `json:"substance"`
	OriginalMg  float64 `json:"original_mg"`
	CurrentMg   float64 `json:"current_mg"` // The calculated value
	TimeElapsed string  `json:"time_elapsed"`
}

// -------------------------------------------------------------------------
// Endpoint 1: Stateless Analysis (POST /analyze)
// -------------------------------------------------------------------------

func (h *Handler) AnalyzeEndpoint(w http.ResponseWriter, r *http.Request) {
	// 1. Parse Request
	var req AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Convert DTO to Domain Model
	var domainStack []domain.ActiveDose
	for _, dto := range req.ActiveStack {
		t, err := time.Parse(time.RFC3339, dto.IngestedAtStr)
		if err != nil {
			http.Error(w, "Invalid time format (use RFC3339): "+dto.IngestedAtStr, http.StatusBadRequest)
			return
		}
		domainStack = append(domainStack, domain.ActiveDose{
			SubstanceID: dto.SubstanceID,
			AmountMg:    dto.AmountMg,
			IngestedAt:  t,
		})
	}

	// 3. Call the Engine
	conflicts, err := h.Advisor.CheckSafety(domainStack, req.ProposedID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 4. Format Response
	type Response struct {
		Safe      bool              `json:"safe"`
		Conflicts []engine.Conflict `json:"conflicts,omitempty"`
	}

	resp := Response{
		Safe:      len(conflicts) == 0,
		Conflicts: conflicts,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// -------------------------------------------------------------------------
// Endpoint 2: Ingest Dose (POST /ingest)
// -------------------------------------------------------------------------

func (h *Handler) IngestEndpoint(w http.ResponseWriter, r *http.Request) {
	var req IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create the Domain Object
	dose := domain.ActiveDose{
		ID:          uuid.New().String(),
		SubstanceID: req.SubstanceID,
		AmountMg:    req.AmountMg,
		IngestedAt:  time.Now(),
	}

	// Save to Store
	h.Store.AddDose(req.UserID, dose)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status": "ingested", "message": "Dose tracked successfully"}`))
}

// -------------------------------------------------------------------------
// Endpoint 3: Check Status (GET /status)
// -------------------------------------------------------------------------

func (h *Handler) StatusEndpoint(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	// 1. Get the raw stack
	stack := h.Store.GetStack(userID)
	var response []StatusResponse
	now := time.Now()

	// 2. Iterate and Calculate Decay
	for _, dose := range stack {
		// Fetch scientific data (Half-Life)
		def, err := h.Repo.GetDefinition(dose.SubstanceID)
		if err != nil {
			continue // Skip unknown substances
		}

		elapsed := now.Sub(dose.IngestedAt)

		// THE MATH: Calculate remaining amount
		remaining := h.Calc.RemainingAmount(dose.AmountMg, def.HalfLifeHours, elapsed)

		response = append(response, StatusResponse{
			Substance:   def.Name,
			OriginalMg:  dose.AmountMg,
			CurrentMg:   remaining,
			TimeElapsed: elapsed.Round(time.Minute).String(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
