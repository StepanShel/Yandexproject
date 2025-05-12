package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/StepanShel/YandexProject/internal/repo"
	parser "github.com/StepanShel/YandexProject/pkg/orchestrator/parser"
	"github.com/StepanShel/YandexProject/proto/calc"
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
	case Expression:
		resp = map[string]Expression{"Expression": data}
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

	username, ok := r.Context().Value("username").(string)
	if !ok {
		respJson(w, errors.New("unauthorized"), http.StatusUnauthorized)
		return
	}

	var request Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respJson(w, errors.New("invalid data"), 422)
		return
	}
	defer r.Body.Close()

	id := uuid.New()

	expr := &repo.Expression{
		ID:         id,
		Username:   username,
		Expression: request.Expression,
		Status:     "processing",
	}

	if err := server.Repo.CreateExpression(expr); err != nil {
		respJson(w, errors.New("failed to save expression"), http.StatusInternalServerError)
		return
	}

	if err := respJson(w, id.String(), 201); err != nil {
		fmt.Println(err)
	}

	go func() {
		fmt.Println("start parsing")
		err := server.startParsingExpression(request.Expression, id)
		if err != nil {
			if updateErr := server.Repo.UpdateExpressionResult(id, 0, "error"); updateErr != nil {
				fmt.Println("failed to update expression status:", updateErr)
			}
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

	username, ok := r.Context().Value("username").(string)
	if !ok {
		respJson(w, errors.New("unauthorized"), http.StatusUnauthorized)
		return
	}

	expressions, err := server.Repo.GetExpressions(username)
	if err != nil {
		fmt.Println(err)
		respJson(w, errors.New("failed to get expressions"), http.StatusInternalServerError)
		return
	}

	var result []Expression
	for _, expr := range expressions {
		result = append(result, Expression{
			ID:     expr.ID.String(),
			Result: float64(expr.Result),
			Status: expr.Status,
		})
	}

	respJson(w, result, 200)
}

// endpoint api/v1/expressions/:id
func (server *Server) HandleExpressionsById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respJson(w, errors.New("unsupported method"), 405)
		return
	}

	username, ok := r.Context().Value("username").(string)
	if !ok {
		respJson(w, errors.New("unauthorized"), http.StatusUnauthorized)
		return
	}

	path := strings.Split(r.URL.Path, "/")
	id, _ := uuid.Parse(path[4])

	expr, err := server.Repo.GetExpressionByID(id)
	if err != nil {
		respJson(w, errors.New("expression not found"), 404)
		return
	}

	if expr.Username != username {
		respJson(w, errors.New("access denied"), http.StatusForbidden)
		return
	}

	respJson(w, Expression{
		ID:     expr.ID.String(),
		Result: float64(expr.Result),
		Status: expr.Status,
	}, 200)
}

func (server *Server) startParsingExpression(expression string, id uuid.UUID) error {
	tasksch := make(chan *calc.Task, 100)
	resultch := make(chan *calc.Result, 100)
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
	var result float64
	var parseErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		result, parseErr = parser.ParsingAST(node, server.Config, tasksch, resultch)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for task := range tasksch {
			server.grpcServer.Tasks <- task
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case resultch <- <-server.grpcServer.AgentCh:
			}
		}
	}()

	wg.Wait()

	if parseErr != nil {
		if err := server.Repo.UpdateExpressionResult(id, 0, "error"); err != nil {
			return fmt.Errorf("update status error: %v, original error: %v", err, parseErr)
		}
		return parseErr
	}
	if err := server.Repo.UpdateExpressionResult(id, int(result), "DONE"); err != nil {
		return err
	}

	return nil
}
