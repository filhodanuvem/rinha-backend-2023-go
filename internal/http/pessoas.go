package http

import (
	"context"
	"encoding/json"
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

	var p rinha.Pessoa
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "expected json body", http.StatusBadRequest)
		return
	}

	if p.Apelido == "" ||
		p.Nome == "" ||
		p.Nascimento.IsZero() {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}
	p.ID = uuid.New()

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if err := repo.Create(ctx, p); err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err == rinha.ErrApelidoJaExiste {
			w.Write([]byte("apelido já existe"))
		}

		return
	}

	w.WriteHeader(http.StatusCreated)
}
