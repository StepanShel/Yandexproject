package repo

import (
	"github.com/google/uuid"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Expression struct {
	ID         uuid.UUID `json:"id"`
	Username   string    `json:"username"`
	Expression string    `json:"expression"`
	Result     int       `json:"result"`
	Status     string    `json:"status"`
	CreatedAt  string    `json:"created_at"`
}
