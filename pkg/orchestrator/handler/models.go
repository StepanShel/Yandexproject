package handler

import (
	"sync"

	"github.com/StepanShel/YandexProject/pkg/orchestrator/config"
	"github.com/StepanShel/YandexProject/pkg/orchestrator/parser"
)

type Request struct {
	Expression string `json:"expression"`
}

type ResponseError struct {
	Error string `json:"error"`
}

type ResponseID struct {
	Id string `json:"id"`
}

type ResultFromAgent struct {
	Result float64 `json:"result"`
	ID     string  `json:"id"`
}

type ResponseExprs struct {
	Exprs []Expression `json:"expressions"`
}

type Expression struct {
	ID     string   `json:"id"`
	Status string   `json:"status"`
	Result *float64 `json:"result"`
}

type Server struct {
	mu          sync.Mutex
	expressions map[string]*Expression
	tasks       []parser.Task
	Agentch     chan parser.Result
	Config      *config.Config
}

func NewServer() *Server {
	return &Server{
		mu:          sync.Mutex{},
		expressions: make(map[string]*Expression),
		tasks:       make([]parser.Task, 0),
		Config:      config.ConfigFromEnv(),
		Agentch:     make(chan parser.Result, 100),
	}
}
func (server *Server) Shutdown() {
	close(server.Agentch)
}
