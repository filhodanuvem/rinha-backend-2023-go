package rinha

import (
	"time"

	"github.com/google/uuid"
)

type Pessoa struct {
	ID         uuid.UUID `json:"id"`
	Apelido    string    `json:"apelido"`
	Nome       string    `json:"nome"`
	Nascimento time.Time `json:"nascimento"`
	Stack      []string  `json:"stack"`
}
