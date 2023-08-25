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

var Repo *Repository

func NewRepository(Conn *pgxpool.Pool, rds *redis.Client) *Repository {
	if Repo == nil {
		Repo = &Repository{Conn: Conn, Cache: rds}
	}

	return Repo
}

func (r *Repository) Create(ctx context.Context, pessoa rinha.Pessoa) error {
	if v, _ := r.Cache.Get(ctx, pessoa.Apelido).Result(); v != "" {
		return rinha.ErrApelidoJaExiste
	}

	j, err := json.Marshal(pessoa)
	if err != nil {
		return err
	}

	pipe := r.Cache.Pipeline()
	pipe.Set(ctx, pessoa.Apelido, "t", 0)
	pipe.Set(ctx, pessoa.ID.String(), j, 24*time.Hour)
	if _, err := pipe.Exec(ctx); err != nil {
		slog.Error(err.Error())
		return err
	}

	go func() {
		index := fmt.Sprintf("%s %s %s", strings.ToLower(pessoa.Apelido), strings.ToLower(pessoa.Nome), strings.ToLower(strings.Join(pessoa.Stack, " ")))
		_, err := r.Conn.Exec(ctx, `
		INSERT INTO pessoas (id, apelido, nome, nascimento, stack, search_index)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, pessoa.ID, pessoa.Apelido, pessoa.Nome, pessoa.Nascimento.Format(time.RFC3339), pessoa.Stack, index)

		if pgerr, ok := err.(*pgconn.PgError); ok {
			if pgerr.ConstraintName == "pessoas_apelido_key" {
				slog.Error(pgerr.Error())
			}
		}
	}()

	return nil
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
	if err != nil && err != redis.Nil {
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
		LIMIT 1
	`, id).Scan(&pessoa.ID, &pessoa.Apelido, &pessoa.Nome, &nascimento, &pessoa.Stack)

	pessoa.Nascimento = rinha.Date{Time: nascimento}

	if err == pgx.ErrNoRows {
		return pessoa, nil
	}

	return pessoa, err
}

func (r *Repository) FindByTermo(ctx context.Context, t string) ([]rinha.Pessoa, error) {
	pessoas := []rinha.Pessoa{}

	rows, err := r.Conn.Query(ctx, `
		SELECT distinct id, apelido, nome, nascimento, stack
		FROM pessoas
		WHERE search_index ILIKE '%' || $1 || '%'
		LIMIT 50
	`, strings.ToLower(t))

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
