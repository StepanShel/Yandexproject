package grpc

import (
	"context"
	"time"

	"github.com/StepanShel/YandexProject/proto/calc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client calc.CalculatorClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: calc.NewCalculatorClient(conn),
	}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) GetTask() (*calc.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	return c.client.GetTask(ctx, &calc.Empty{})
}

func (c *Client) SendResult(taskID string, result float64, err error) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	res := &calc.Result{
		TaskId: taskID,
		Result: float32(result),
	}
	if err != nil {
		res.Error = err.Error()
	}

	_, err = c.client.SendResult(ctx, res)
	return err
}
