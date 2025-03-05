package main

import (
	"log"
	"net/http"

	"github.com/sklerakuku/calc3.0/internal/config"
	"github.com/sklerakuku/calc3.0/internal/orchestator"
)

func main() {
	cfg := config.Load()
	orch := orchestator.New(cfg)

	http.HandleFunc("/", orchestator.ServeHTML)
	http.HandleFunc("/api/v1/calculate", orch.AddCalculation)
	http.HandleFunc("/api/v1/expressions", orch.GetExpressions)
	http.HandleFunc("/api/v1/expressions/", orch.GetExpression)
	http.HandleFunc("/internal/task", orch.HandleTask)

	log.Printf("Orchestrator starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
