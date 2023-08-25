package config

import (
	"os"
	"time"
)

var DatabaseURL string
var CacheURL string
var PROFILING bool
var NumBatch = 100
var NumWorkers = 10
var WorkerTimeout = 500 * time.Millisecond

func init() {
	DatabaseURL = envOrFatal("DATABASE_URL")
	CacheURL = envOrFatal("CACHE_URL")
	PROFILING = os.Getenv("PROFILING") == "true"
}

func envOrFatal(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("missing required environment variable " + key)
	}

	return value
}
