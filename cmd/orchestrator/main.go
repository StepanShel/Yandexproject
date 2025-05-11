package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/StepanShel/YandexProject/internal/auth"
	"github.com/StepanShel/YandexProject/pkg/orchestrator/handler"
)

func main() {
	server := handler.NewServer()
	jwtService := auth.NewJWTService("secret")
	authHandler := auth.NewAuthHandler(server.Repo, jwtService)

	http.HandleFunc("/api/v1/register", authHandler.Register)
	http.HandleFunc("/api/v1/login", authHandler.Login)

	http.HandleFunc("/api/v1/calculate", jwtService.AuthMiddleware(server.HandleCalculate))
	http.HandleFunc("/api/v1/expressions", jwtService.AuthMiddleware(server.HandleExpressions))
	http.HandleFunc("/api/v1/expressions/{id}", jwtService.AuthMiddleware(server.HandleExpressionsById))
	http.HandleFunc("GET /internal/task", server.HandleTaskGet)
	http.HandleFunc("POST /internal/task", server.HandleTaskPost)

	fmt.Printf("Orchestrator is running on http://localhost:%s\n", server.Config.Port)
	addr := fmt.Sprintf(":%s", server.Config.Port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	fmt.Println("Server stopped")
}
