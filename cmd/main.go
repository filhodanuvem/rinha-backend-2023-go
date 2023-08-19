package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"runtime/pprof"
	"runtime/trace"

	"github.com/filhodanuvem/rinha/internal/cache"
	"github.com/filhodanuvem/rinha/internal/config"
	"github.com/filhodanuvem/rinha/internal/database"
	route "github.com/filhodanuvem/rinha/internal/http"
)

func main() {
	if config.PROFILING {
		slog.Info("Running with profiling")
		// Create a CPU profile file
		f, err := os.Create("/pprof/profile.prof")
		if err != nil {
			panic(err)
		}
		defer f.Close()

		// Start CPU profiling
		if err := pprof.StartCPUProfile(f); err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()

		// Start tracing
		traceFile, err := os.Create("/pprof/trace.out")
		if err != nil {
			panic(err)
		}
		defer traceFile.Close()

		if err := trace.Start(traceFile); err != nil {
			panic(err)
		}
		defer trace.Stop()
	}

	slog.Info("Waiting for database...")
	time.Sleep(5 * time.Second) // @TODO wait for database on docker-compose
	if err := database.Connect(); err != nil {
		panic(err)
	}
	defer database.Close()

	if err := cache.Connect(); err != nil {
		panic(err)
	}

	// http.HandleFunc("/pessoas/", route.Pessoas)
	http.HandleFunc("/", route.Pessoas)

	slog.Info("Server running on port 80")
	go http.ListenAndServe(":80", nil)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
}
