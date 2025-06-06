package handler

import (
	"fmt"
	"sync"

	"github.com/StepanShel/YandexProject/internal/repo"
	"github.com/StepanShel/YandexProject/pkg/orchestrator/config"
	grpc "github.com/StepanShel/YandexProject/pkg/orchestrator/gRPC"
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
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

type Server struct {
	grpcServer *grpc.Server
	Repo       *repo.Repo
	mu         sync.Mutex
	tasks      []parser.Task
	Agentch    chan parser.Result
	Config     *config.Config
}

func NewServer(grpcServer *grpc.Server) *Server {
	Repo, err := repo.NewRepository()
	if err != nil {
		fmt.Printf("failed to init repository: %v", err)
		return nil
	}

	return &Server{
		grpcServer: grpcServer,
		Repo:       Repo,
		tasks:      make([]parser.Task, 0),
		Config:     config.ConfigFromEnv(),
	}
}
