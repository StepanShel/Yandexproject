package agent

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	grpc "github.com/StepanShel/YandexProject/internal/agent/gRPC"
)

func NewAgent(client *grpc.Client) *Agent {
	compPower, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if compPower == 0 {
		compPower = 1
	}

	return &Agent{
		compPower: compPower,
		client:    client,
	}
}

func (a *Agent) StartWorkers() {
	for i := 0; i < a.compPower; i++ {
		go a.Worker(i)
	}
}

func (a *Agent) Worker(id int) {
	for {
		task, err := a.client.GetTask()
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}

		log.Printf("Worker %d received task: %f %s %f", id, task.Arg1, task.Operation, task.Arg2)

		result, err := Calculate(task.Operation, int(task.OperationTime), float64(task.Arg1), float64(task.Arg2))
		if err != nil {
			log.Printf("Worker %d: calculation error: %v", id, err)
			continue
		}

		if err := a.client.SendResult(task.Id, result, nil); err != nil {
			log.Printf("Worker %d: failed to send result: %v", id, err)
			continue
		}

		log.Printf("Worker %d sent result: %f for task %s", id, result, task.Id)
	}
}

func Calculate(operation string, duration int, a, b float64) (float64, error) {
	switch operation {
	case "+":
		time.Sleep(time.Millisecond * time.Duration(duration))
		return a + b, nil
	case "-":
		time.Sleep(time.Millisecond * time.Duration(duration))
		return a - b, nil
	case "*":
		time.Sleep(time.Millisecond * time.Duration(duration))
		return a * b, nil
	case "/":
		time.Sleep(time.Millisecond * time.Duration(duration))
		return a / b, nil
	default:
		return 0, fmt.Errorf("invalid operator: %s", operation)
	}
}
