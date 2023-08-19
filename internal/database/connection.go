package database

import (
	"context"

	"github.com/filhodanuvem/rinha/internal/config"
	"github.com/jackc/pgx/v4/pgxpool"
)

var Connection *pgxpool.Pool

func Connect() error {
	var err error
	Connection, err = pgxpool.Connect(context.Background(), config.DatabaseURL)

	return err
}

func Close() {
	if Connection == nil {
		return
	}
	Connection.Close()
}
