package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/StepanShel/YandexProject/internal/auth"
	GRPC "github.com/StepanShel/YandexProject/pkg/orchestrator/gRPC"
	"github.com/StepanShel/YandexProject/pkg/orchestrator/handler"
	"github.com/StepanShel/YandexProject/proto/calc"
	"google.golang.org/grpc"
)

func main() {
	grpcServer := grpc.NewServer()
	calcService := GRPC.NewServer()
	calc.RegisterCalculatorServer(grpcServer, calcService)
	go func() {
		log.Printf("gRPC server listening on :5000")
		lis, err := net.Listen("tcp", ":5000")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	server := handler.NewServer(calcService)
	jwtService := auth.NewJWTService("secret")
	authHandler := auth.NewAuthHandler(server.Repo, jwtService)

	http.HandleFunc("/api/v1/register", authHandler.Register)
	http.HandleFunc("/api/v1/login", authHandler.Login)

	http.HandleFunc("/api/v1/calculate", jwtService.AuthMiddleware(server.HandleCalculate))
	http.HandleFunc("/api/v1/expressions", jwtService.AuthMiddleware(server.HandleExpressions))
	http.HandleFunc("/api/v1/expressions/{id}", jwtService.AuthMiddleware(server.HandleExpressionsById))

	fmt.Printf("Orchestrator is running on http://localhost:%s\n", server.Config.Port)
	addr := fmt.Sprintf(":%s", server.Config.Port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	fmt.Println("Server stopped")
}
