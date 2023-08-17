package pessoa

import (
	"context"
	"encoding/json"

	"log/slog"

	"github.com/filhodanuvem/rinha"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	Conn  *pgxpool.Pool
	Cache *redis.Client
}

func (r *Repository) Create(ctx context.Context, pessoa rinha.Pessoa) error {
	_, err := r.Conn.Exec(ctx, `
		INSERT INTO pessoas (id, apelido, nome, nascimento, stack)
		VALUES ($1, $2, $3, $4, $5)
	`, pessoa.ID, pessoa.Apelido, pessoa.Nome, pessoa.Nascimento, pessoa.Stack)

	if err == nil {
		go func() {
			if _, err := r.Cache.Set(ctx, pessoa.ID.String(), pessoa, 0).Result(); err != nil {
				slog.Error(err.Error())
			}
		}()
	}

	if pgerr, ok := err.(*pgconn.PgError); ok {
		if pgerr.ConstraintName == "pessoas_apelido_key" {
			return rinha.ErrApelidoJaExiste
		}
	}

	return err
}

func (r *Repository) Find(ctx context.Context, apelido string) (rinha.Pessoa, error) {
	var pessoa rinha.Pessoa

	entry, err := r.Cache.Get(ctx, apelido).Result()
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

	err = r.Conn.QueryRow(ctx, `
		SELECT id, apelido, nome, nascimento, stack
		FROM pessoas
		WHERE apelido = $1
	`, apelido).Scan(&pessoa.ID, &pessoa.Apelido, &pessoa.Nome, &pessoa.Nascimento, &pessoa.Stack)

	return pessoa, err
}
