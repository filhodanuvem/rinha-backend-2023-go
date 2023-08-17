package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/filhodanuvem/rinha/internal/cache"
	"github.com/filhodanuvem/rinha/internal/database"
	route "github.com/filhodanuvem/rinha/internal/http"
)

func main() {
	slog.Info("Waiting for database...")
	time.Sleep(5 * time.Second) // @TODO wait for database on docker-compose
	if err := database.Connect(); err != nil {
		panic(err)
	}
	if err := cache.Connect(); err != nil {
		panic(err)
	}

	http.HandleFunc("/pessoas/", route.Pessoas)
	http.HandleFunc("/pessoas", route.Pessoas)

	slog.Info("Server running on port 80")
	http.ListenAndServe(":80", nil)
}
