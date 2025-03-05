package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/StepanShel/YandexProject/pkg/orchestrator/config"
	parser "github.com/StepanShel/YandexProject/pkg/orchestrator/parser"
)

func TestHandleCalculate(t *testing.T) {
	server := &Server{
		expressions: make(map[string]*Expression),
		tasks:       make([]parser.Task, 0),
		Config: &config.Config{
			AddTime:       200,
			Subtime:       200,
			MultiplicTime: 200,
			Divtime:       200,
		},
	}

	requestBody := `{"expression": "2+2"}`
	req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleCalculate)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var response ResponseID
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response.Id == "" {
		t.Error("handler returned empty ID")
	}

}
func TestHandleExpressions(t *testing.T) {
	server := &Server{
		expressions: make(map[string]*Expression),
		tasks:       make([]parser.Task, 0),
		Config: &config.Config{
			AddTime:       200,
			Subtime:       200,
			MultiplicTime: 200,
			Divtime:       200,
		},
	}

	requestBody := `{"expression": "2+2"}`
	req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(requestBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleCalculate)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var response ResponseID
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if response.Id == "" {
		t.Error("Handler returned empty expression ID")
	}

	server.mu.Lock()
	expr, exists := server.expressions[response.Id]
	server.mu.Unlock()

	if !exists {
		t.Errorf("Expression with ID %v not found in server.expressions", response.Id)
	}

	if expr.Status != "processing" {
		t.Errorf("Handler returned wrong expression status: got %v want %v", expr.Status, "processing")
	}
}

func TestHandleExpressions_Get(t *testing.T) {
	server := &Server{
		expressions: map[string]*Expression{
			"1": {ID: "1", Status: "processing"},
			"2": {ID: "2", Status: "DONE", Result: func() *float64 { res := 42.0; return &res }()},
		},
		mu: sync.Mutex{},
	}

	req, err := http.NewRequest("GET", "/api/v1/expressions", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleExpressions)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response struct {
		Expressions []Expression `json:"expressions"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if len(response.Expressions) != 2 {
		t.Errorf("Handler returned wrong number of expressions: got %v want %v", len(response.Expressions), 2)
	}

	if response.Expressions[0].ID != "1" || response.Expressions[0].Status != "processing" || response.Expressions[0].Result != nil {
		t.Errorf("Handler returned wrong first expression: got %+v", response.Expressions[0])
	}

	if response.Expressions[1].ID != "2" || response.Expressions[1].Status != "DONE" || *response.Expressions[1].Result != 42.0 {
		t.Errorf("Handler returned wrong second expression: got %+v", response.Expressions[1])
	}
}

func TestHandleExpressionsById_NotFound(t *testing.T) {
	server := &Server{
		expressions: map[string]*Expression{
			"1": {ID: "1", Status: "processing"},
		},
		mu: sync.Mutex{},
	}

	req, err := http.NewRequest("GET", "/api/v1/expressions/2", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleExpressionsById)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var response ResponseError
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if response.Error != "expression not found" {
		t.Errorf("Handler returned wrong error message: got %v want %v", response.Error, "expression not found")
	}
}
