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
	Conn      *pgxpool.Pool
	Cache     *redis.Client
	ChPessoas chan rinha.Pessoa
}

var Repo *Repository

func NewRepository(Conn *pgxpool.Pool, c *redis.Client) *Repository {
	if Repo == nil {
		Repo = &Repository{Conn: Conn, ChPessoas: make(chan rinha.Pessoa, 1000), Cache: c}
	}

	return Repo
}

func (r *Repository) Insert(pessoas []rinha.Pessoa) error {
	if len(pessoas) == 0 {
		return nil
	}
	params := make([]interface{}, 0, len(pessoas)*5)
	values := ""
	j := 0
	for i, p := range pessoas {
		params = append(params, p.ID, p.Apelido, p.Nome, p.Nascimento.Format(time.RFC3339), p.Stack)

		values += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", j+1, j+2, j+3, j+4, j+5)
		if i != len(pessoas)-1 {
			values += ","
		}
		j += 5
	}

	_, err := r.Conn.Exec(context.Background(), fmt.Sprintf(`
		INSERT INTO pessoas (id, apelido, nome, nascimento, stack)
		VALUES %s
	`, values), params...)

	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.ConstraintName == "pessoas_apelido_key" {
		// @TODO how to deal with conflicts on database
		slog.Error("algum apelido ja existe")
		return pgErr
	}

	return err
}

func (r *Repository) Create(ctx context.Context, pessoa rinha.Pessoa) error {
	if v, _ := r.Cache.Get(ctx, pessoa.Apelido).Result(); v != "" {
		return rinha.ErrApelidoJaExiste
	}

	r.ChPessoas <- pessoa

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

	var nascimento time.Time
	err := r.Conn.QueryRow(ctx, `
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
		WHERE apelido ILIKE '%' || $1 || '%' OR 
			nome ILIKE '%' || $1 || '%'
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
