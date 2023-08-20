package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/filhodanuvem/rinha"
	"github.com/filhodanuvem/rinha/internal/pessoa"
	"github.com/google/uuid"
)

func CountPessoas(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	count, err := pessoa.Repo.Count(ctx)
	if err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := strconv.Itoa(count)
	w.Write([]byte(response))
}

func Pessoas(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.URL.Path == "/pessoas" {
		PostPessoas(w, r)
		return
	}

	if r.Method == http.MethodGet && strings.Index(r.URL.Path, "/pessoas") == 0 {
		GetPessoas(w, r)
		return
	}

	if r.Method == http.MethodGet && r.URL.Path == "/contagem-pessoas" {
		CountPessoas(w, r)
		return
	}

	w.Header().Set("Allow", "GET,POST")
	w.WriteHeader(http.StatusNotFound)
}

func PostPessoas(w http.ResponseWriter, r *http.Request) {
	var p rinha.Pessoa
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		slog.Debug(err.Error())
		http.Error(w, "expected json body", http.StatusBadRequest)
		return
	}

	if p.Apelido == "" ||
		p.Nome == "" ||
		p.Nascimento.IsZero() {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	if utf8.RuneCountInString(p.Apelido) > 32 ||
		utf8.RuneCountInString(p.Nome) > 32 {
		http.Error(w, "fields too long", http.StatusBadRequest)
		return
	}

	p.ID = uuid.New()

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	if err := pessoa.Repo.Create(ctx, p); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err == rinha.ErrApelidoJaExiste {
			w.Write([]byte("apelido já existe"))
			return
		}
		slog.Error(err.Error())

		return
	}

	w.Header().Set("Location", "/pessoas/"+p.ID.String())

	j, err := json.Marshal(p)
	if err != nil {
		slog.Debug(err.Error())
		w.WriteHeader(http.StatusCreated)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(j)
}

func GetPessoas(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	var id string
	if strings.Contains(path, "/pessoas/") {
		id = path[len("/pessoas/"):]
	}

	if id != "" {
		GetPessoaByID(w, r, id)
		return
	}

	GetPessoasByTermo(w, r)
}

func GetPessoaByID(w http.ResponseWriter, r *http.Request, param string) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	id, err := uuid.Parse(param)
	if err != nil {
		slog.Debug(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	pessoa, err := pessoa.Repo.FindOne(ctx, id)
	if err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if pessoa.ID == uuid.Nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	j, err := json.Marshal(pessoa)
	if err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func GetPessoasByTermo(w http.ResponseWriter, r *http.Request) {
	termo := r.URL.Query().Get("t")
	if termo == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing required query param 't'"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	pessoas, err := pessoa.Repo.FindByTermo(ctx, termo)
	if err != nil {
		slog.Error("error when finding by t:" + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err := json.Marshal(pessoas)
	if err != nil {
		slog.Error("error when creating response collection:" + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
