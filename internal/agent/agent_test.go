package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		operation string
		duration  int
		a, b      float64
		expected  float64
		err       error
	}{
		{"+", 100, 2, 3, 5, nil},
		{"-", 100, 5, 3, 2, nil},
		{"*", 100, 2, 3, 6, nil},
		{"/", 100, 6, 3, 2, nil},
		{"%", 100, 6, 3, 0, fmt.Errorf("invalid operator: %%")},
	}

	for _, test := range tests {
		result, err := Calculate(test.operation, test.duration, test.a, test.b)
		if err != nil && test.err == nil {
			t.Errorf("Calculate(%v, %v, %v, %v) returned unexpected error: %v", test.operation, test.duration, test.a, test.b, err)
		}
		if err == nil && test.err != nil {
			t.Errorf("Calculate(%v, %v, %v, %v) expected error: %v", test.operation, test.duration, test.a, test.b, test.err)
		}
		if result != test.expected {
			t.Errorf("Calculate(%v, %v, %v, %v) = %v, expected %v", test.operation, test.duration, test.a, test.b, result, test.expected)
		}
	}
}
func TestWorker(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/task" && r.Method == http.MethodGet {
			task := struct {
				AgTask `json:"task"`
			}{
				AgTask: AgTask{
					ID:            "1",
					Arg1:          2,
					Arg2:          3,
					Operation:     "+",
					OperationTime: 100,
				},
			}
			json.NewEncoder(w).Encode(task)
			return
		}

		if r.URL.Path == "/internal/task" && r.Method == http.MethodPost {
			var response Response
			if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response body: %v", err)
			}
			if response.Res != 5 {
				t.Errorf("Expected result 5, got %v", response.Res)
			}
			if response.Id != "1" {
				t.Errorf("Expected task ID 1, got %v", response.Id)
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	agent := &Agent{
		compPower: 1,
		port:      8081,
	}

	go agent.Worker(0)

	time.Sleep(time.Second * 2)
}
