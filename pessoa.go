package rinha

import (
	"time"

	"github.com/google/uuid"
)

type Pessoa struct {
	ID         uuid.UUID
	Apelido    string
	Nome       string
	Nascimento time.Time
	Stack      []string
}
