package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration parameters.
type Config struct {
	Port             string
	DBDriver         string
	DBSource         string
	WorkerCount      int
	BatchTimeout     int
	ConcurrencyLimit int
	EnableMetrics    bool
}

// LoadConfig retrieves configuration from environment variables with sensible defaults.
func LoadConfig() *Config {
	port := getEnv("PORT", "8080")
	dbDriver := getEnv("DB_DRIVER", "sqlite")
	dbSource := getEnv("DB_SOURCE", ":memory:")
	workerCountStr := getEnv("WORKER_COUNT", "8")
	batchTimeoutStr := getEnv("BATCH_TIMEOUT_SEC", "60")
	concurrencyLimitStr := getEnv("CONCURRENCY_LIMIT", "16")

	workerCount, err := strconv.Atoi(workerCountStr)
	if err != nil {
		workerCount = 8
	}

	batchTimeout, err := strconv.Atoi(batchTimeoutStr)
	if err != nil {
		batchTimeout = 60
	}

	concurrencyLimit, err := strconv.Atoi(concurrencyLimitStr)
	if err != nil {
		concurrencyLimit = 16
	}

	return &Config{
		Port:             port,
		DBDriver:         dbDriver,
		DBSource:         dbSource,
		WorkerCount:      workerCount,
		BatchTimeout:     batchTimeout,
		ConcurrencyLimit: concurrencyLimit,
		EnableMetrics:    getEnv("ENABLE_METRICS", "true") == "true",
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
