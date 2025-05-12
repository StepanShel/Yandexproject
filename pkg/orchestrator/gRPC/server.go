package grpc

import (
	"context"

	"github.com/StepanShel/YandexProject/proto/calc"
)

type Server struct {
	calc.CalculatorServer
	Tasks   chan *calc.Task
	AgentCh chan *calc.Result
}

func NewServer() *Server {
	return &Server{
		Tasks: make(chan *calc.Task, 100),
	}
}

func (s *Server) GetTask(ctx context.Context, _ *calc.Empty) (*calc.Task, error) {
	select {
	case task := <-s.Tasks:
		return task, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *Server) SendResult(ctx context.Context, result *calc.Result) (*calc.Empty, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.AgentCh <- result:
		return &calc.Empty{}, nil
	}
}
