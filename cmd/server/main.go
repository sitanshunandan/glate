package main

import (
	"log"
	"net/http"
	"time"

	"github.com/sitanshunandan/glate/internal/api"
	"github.com/sitanshunandan/glate/internal/engine"
	"github.com/sitanshunandan/glate/internal/repository"
	"github.com/sitanshunandan/glate/internal/store"
)

func main() {
	// 1. Dependencies
	repo, err := repository.NewInMemoryRepo("configs/substances.json")
	if err != nil {
		log.Fatalf("Config Error: %v", err)
	}

	calc := engine.NewMetabolicCalculator()
	advisor := engine.NewAdvisor(repo, calc)
	sessionStore := store.NewSessionStore()

	handler := api.NewHandler(advisor, sessionStore, repo, calc)

	// 2. Start the Background Monitor (NEW)
	// We set it to run every 10 seconds for the demo.
	monitor := engine.NewMonitor(sessionStore, repo, calc)
	monitor.Start(10 * time.Second)

	// 3. Router
	mux := http.NewServeMux()
	mux.HandleFunc("POST /analyze", handler.AnalyzeEndpoint)
	mux.HandleFunc("POST /ingest", handler.IngestEndpoint)
	mux.HandleFunc("GET /status", handler.StatusEndpoint)

	// 4. Server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("🚀 Glate Stateful Server listening on :8080...")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
