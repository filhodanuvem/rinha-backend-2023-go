package config

import (
	"os"
	"runtime"
	"time"
)

var DatabaseURL string
var CacheURL string
var PROFILING bool
var NumBatch = 100
var NumWorkers = 1
var WorkerTimeout = 1 * time.Second

func init() {
	NumWorkers = runtime.GOMAXPROCS(0) * 2
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
