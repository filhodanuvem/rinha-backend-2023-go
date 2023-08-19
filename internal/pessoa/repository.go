package pessoa

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"log/slog"

	"github.com/filhodanuvem/rinha"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	Conn  *pgxpool.Pool
	Cache *redis.Client
}

func (r *Repository) Create(ctx context.Context, pessoa rinha.Pessoa) error {
	if v, _ := r.Cache.Get(ctx, pessoa.Apelido).Result(); v != "" {
		return rinha.ErrApelidoJaExiste
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		index := fmt.Sprintf("%s %s %s", strings.ToLower(pessoa.Apelido), strings.ToLower(pessoa.Nome), strings.ToLower(strings.Join(pessoa.Stack, " ")))
		_, err := r.Conn.Exec(ctx, `
		INSERT INTO pessoas (id, apelido, nome, nascimento, stack, search_index)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, pessoa.ID, pessoa.Apelido, pessoa.Nome, pessoa.Nascimento.Format(time.RFC3339), pessoa.Stack, index)

		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.ConstraintName == "pessoas_apelido_key" {
			r.Cache.Set(ctx, pessoa.Apelido, "t", 0)
			return
		}

		if err != nil {
			slog.Error(err.Error())
		}
	}()

	j, err := json.Marshal(pessoa)
	if err != nil {
		slog.Error(err.Error())
	}

	if _, err := r.Cache.Set(ctx, pessoa.ID.String(), j, 24*time.Hour).Result(); err != nil {
		slog.Error(err.Error())
	}
	r.Cache.Set(ctx, pessoa.Apelido, "t", 0)

	return err
}

func (r *Repository) Count(ctx context.Context) (int, error) {
	var count int

	err := r.Conn.QueryRow(ctx, `
		SELECT COUNT(id)
		FROM pessoas
	`).Scan(&count)

	return count, err
}

func (r *Repository) FindOne(ctx context.Context, id uuid.UUID) (rinha.Pessoa, error) {
	var pessoa rinha.Pessoa

	entry, err := r.Cache.Get(ctx, id.String()).Result()
	if err != nil {
		slog.Error(err.Error())
	}

	if entry != "" {
		err := json.Unmarshal([]byte(entry), &pessoa)
		if err == nil {
			return pessoa, nil
		}

		slog.Error(err.Error())
	}

	var nascimento time.Time
	err = r.Conn.QueryRow(ctx, `
		SELECT id, apelido, nome, nascimento, stack
		FROM pessoas
		WHERE id = $1
		LIMIT 1x
	`, id).Scan(&pessoa.ID, &pessoa.Apelido, &pessoa.Nome, &nascimento, &pessoa.Stack)

	pessoa.Nascimento = rinha.Date{Time: nascimento}

	if err == pgx.ErrNoRows {
		return pessoa, nil
	}

	return pessoa, err
}

func (r *Repository) FindByTermo(ctx context.Context, termo string) ([]rinha.Pessoa, error) {
	pessoas := []rinha.Pessoa{}

	rows, err := r.Conn.Query(ctx, `
		SELECT distinct id, apelido, nome, nascimento, stack
		FROM pessoas
		WHERE search_index LIKE '%' || $1 || '%'
		LIMIT 50
	`, termo)

	if err != nil {
		return pessoas, err
	}

	defer rows.Close()

	for rows.Next() {
		var pessoa rinha.Pessoa
		var nascimento time.Time
		err := rows.Scan(&pessoa.ID, &pessoa.Apelido, &pessoa.Nome, &nascimento, &pessoa.Stack)
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		pessoa.Nascimento = rinha.Date{Time: nascimento}
		pessoas = append(pessoas, pessoa)
	}

	return pessoas, err
}
