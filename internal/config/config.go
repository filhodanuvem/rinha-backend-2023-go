package config

import (
	"os"
	"time"
)

var DatabaseURL string
var PROFILING bool
var NumBatch = 100
var NumWorkers = 100
var WorkerTimeout = 1 * time.Second

func init() {
	DatabaseURL = envOrFatal("DATABASE_URL")
	PROFILING = os.Getenv("PROFILING") == "true"
}

func envOrFatal(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("missing required environment variable " + key)
	}

	return value
}
