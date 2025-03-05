package main

import (
	"github.com/StepanShel/YandexProject/internal/agent"
)

func main() {
	agent := agent.NewAgent()
	agent.StartWorkers()

	select {}
}
