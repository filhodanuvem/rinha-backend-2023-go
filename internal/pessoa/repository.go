package pessoa

import (
	"context"

	"github.com/filhodanuvem/rinha"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Repository struct {
	Conn *pgxpool.Pool
}

func (r *Repository) Create(ctx context.Context, pessoa rinha.Pessoa) error {
	_, err := r.Conn.Exec(ctx, `
		INSERT INTO pessoas (id, apelido, nome, nascimento, stack)
		VALUES ($1, $2, $3, $4, $5)
	`, pessoa.ID, pessoa.Apelido, pessoa.Nome, pessoa.Nascimento, pessoa.Stack)

	return err
}
