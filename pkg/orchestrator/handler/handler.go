package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	parser "github.com/StepanShel/YandexProject/pkg/orchestrator/parser"
	uuid "github.com/google/uuid"
)

func respJson(w http.ResponseWriter, data any, errCode int) error {
	w.Header().Set("Content-Type", "application/json")

	var resp any

	switch data := data.(type) {
	case error:
		resp = ResponseError{Error: data.Error()}
	case string:
		resp = ResponseID{Id: data}
	case []Expression:
		resp = ResponseExprs{Exprs: data}
	case parser.Task:
		resp = map[string]parser.Task{"task": data}
	case *Expression:
		resp = map[string]*Expression{"Expression": data}
	}

	w.WriteHeader(errCode)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return err
	}

	return nil
}

// endpoint api/v1/calculate
func (server *Server) HandleCalculate(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		respJson(w, errors.New("unsupported method"), 405)
		return
	}
	var request Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		if err := respJson(w, errors.New("invalid data"), 422); err != nil {
			fmt.Println(err)
		}
		return
	}
	defer r.Body.Close()

	id := uuid.New().String()
	server.mu.Lock()
	server.expressions[id] = &Expression{ID: id, Status: "processing"}
	server.mu.Unlock()

	if err := respJson(w, id, 201); err != nil {
		fmt.Println(err)
	}

	go func() {
		fmt.Println("start parsing")
		err := server.startParsingExpression(request.Expression, id)
		if err != nil {
			server.mu.Lock()
			server.expressions[id].Status = "error"
			server.mu.Unlock()
			fmt.Println("parsing failed:", err)
		} else {
			fmt.Println("parsing completed successfully")
		}
	}()
}

// endpoint api/v1/expressions
func (server *Server) HandleExpressions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respJson(w, errors.New("unsupported method"), 405)
		return
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	var expressions = []Expression{}
	for _, expr := range server.expressions {
		expressions = append(expressions, *expr)
	}

	respJson(w, expressions, 200)

}

// endpoint api/v1/expressions/:id
func (server *Server) HandleExpressionsById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		respJson(w, errors.New("unsupported method"), 405)
		return
	}

	server.mu.Lock()
	defer server.mu.Unlock()
	path := strings.Split(r.URL.Path, "/")

	id := path[4]
	exp, ok := server.expressions[id]
	if !ok {
		respJson(w, errors.New("expression not found"), 400)
	}
	respJson(w, exp, 200)
}

// endpoint GET internal/task
func (server *Server) HandleTaskGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respJson(w, errors.New("unsupported method"), 405)
		return
	}
	if len(server.tasks) == 0 {
		respJson(w, errors.New("no tasks available"), 404)
		return
	}
	server.mu.Lock()
	defer server.mu.Unlock()
	task := server.tasks[0]
	server.tasks = server.tasks[1:]

	respJson(w, task, 200)
}

// endpoint POST internal/task
func (server *Server) HandleTaskPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respJson(w, errors.New("unsupported method"), 405)
		return
	}

	var result parser.Result

	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		respJson(w, errors.New("invalid data"), 422)
		return
	}
	defer r.Body.Close()
	server.Agentch <- result
	w.WriteHeader(http.StatusOK)
}

func (server *Server) startParsingExpression(expression, id string) error {
	tasksch := make(chan parser.Task, server.Config.CompPower)
	resultch := make(chan parser.Result, server.Config.CompPower)
	server.Agentch = make(chan parser.Result, server.Config.CompPower)
	defer close(server.Agentch)
	defer close(tasksch)
	defer close(resultch)

	tokens, err := parser.Tokenize(expression)
	if err != nil {
		return err
	}
	node, err := parser.Ast(tokens)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		var res float64
		res, err := parser.ParsingAST(node, server.Config, tasksch, resultch)
		if err != nil {
			server.mu.Lock()
			server.expressions[id].Status = "error"
			server.mu.Unlock()
			return
		}

		server.mu.Lock()
		server.expressions[id].Result = &res
		server.expressions[id].Status = "DONE"
		server.mu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for task := range tasksch {
			if task.Err != nil {
				fmt.Println("Task error:", task.Err)
				server.mu.Lock()
				server.expressions[id].Status = "error"
				server.mu.Unlock()
				return
			}

			server.mu.Lock()
			server.tasks = append(server.tasks, task)
			server.mu.Unlock()
		}
	}()

	wg.Add(1)
	func() {
		defer wg.Done()
		for agentresp := range server.Agentch {
			resultch <- agentresp
		}
	}()

	wg.Wait()

	return nil
}
