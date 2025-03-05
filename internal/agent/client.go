package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

func NewAgent() *Agent {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8081
	}
	compPower, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if compPower == 0 {
		compPower = 1
	}
	return &Agent{
		compPower: compPower,
		port:      port,
	}
}

func (agent *Agent) StartWorkers() {
	for i := 0; i < agent.compPower; i++ {
		go agent.Worker(i)
	}
}

func (agent *Agent) Worker(id int) {
	client := &http.Client{}
	for {
		resp, err := client.Get(fmt.Sprint("http://localhost:", agent.port, "/internal/task"))
		if err != nil {
			time.Sleep(time.Second * 3)
			continue
		}
		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			time.Sleep(time.Second * 3)
			continue
		}
		var taskResp struct {
			AgTask `json:"task"`
		}
		json.NewDecoder(resp.Body).Decode(&taskResp)
		fmt.Println("Worker ", id, "recieved task", taskResp.Arg1, taskResp.Operation, taskResp.Arg2)
		resp.Body.Close()

		result, err := Calculate(taskResp.Operation, taskResp.OperationTime, taskResp.Arg1, taskResp.Arg2)
		if err != nil {
			fmt.Println("err in agent calculate")
			continue
		}

		response := Response{
			Res: result,
			Id:  taskResp.ID,
		}
		buffer, _ := json.Marshal(response)
		respPost, err := client.Post(fmt.Sprint("http://localhost:", agent.port, "/internal/task"), "application/json", bytes.NewReader(buffer))
		fmt.Println("Agent send an ans", response.Res)
		if err != nil {
			fmt.Println("error")
			continue
		}
		if respPost.StatusCode != http.StatusOK {
			fmt.Println("error")
			continue
		}
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
