package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/StepanShel/YandexProject/pkg/orchestrator/handler"
)

func main() {
	server := handler.NewServer()
	http.HandleFunc("/api/v1/calculate", server.HandleCalculate)
	http.HandleFunc("/api/v1/expressions", server.HandleExpressions)
	http.HandleFunc("/api/v1/expressions/{id}", server.HandleExpressionsById)
	http.HandleFunc("GET /internal/task", server.HandleTaskGet)
	http.HandleFunc("POST /internal/task", server.HandleTaskPost)

	fmt.Printf("Orchestrator is running on http://localhost:%s\n", server.Config.Port)
	addr := fmt.Sprintf(":%s", server.Config.Port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	fmt.Println("Server stopped")
}
