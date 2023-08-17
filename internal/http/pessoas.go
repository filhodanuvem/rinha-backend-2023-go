package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/filhodanuvem/rinha"
	"github.com/filhodanuvem/rinha/internal/database"
	"github.com/filhodanuvem/rinha/internal/pessoa"
	"github.com/google/uuid"
)

func Pessoas(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostPessoas(w, r)
		return
	}

	w.Header().Set("Allow", "GET,POST")
	w.WriteHeader(http.StatusMethodNotAllowed)

}

func PostPessoas(w http.ResponseWriter, r *http.Request) {
	repo := pessoa.Repository{Conn: database.Connection}

	p := rinha.Pessoa{
		ID:         uuid.New(),
		Apelido:    "filhodanuvem",
		Nome:       "Filho da Nuvem",
		Nascimento: time.Now(),
		Stack:      []string{"Go", "Python", "JavaScript"},
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if err := repo.Create(ctx, p); err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
