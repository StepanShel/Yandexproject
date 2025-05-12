package main

import (
	"log"

	"github.com/StepanShel/YandexProject/internal/agent"
	grpc "github.com/StepanShel/YandexProject/internal/agent/gRPC"
)

func main() {
	client, err := grpc.NewClient(":5000")
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer client.Close()

	agent := agent.NewAgent(client)
	agent.StartWorkers()

	select {}
}
