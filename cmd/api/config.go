package main

import (
	"os"

	"github.com/research/phase1a/internal/config"
)

func loadConfig() config.RuntimeConfig {
	grpcAddr := envOr("GRPC_ADDR", ":50051")
	redisAddr := envOr("REDIS_ADDR", "localhost:6379")
	redisPassword := envOr("REDIS_PASSWORD", "")
	redisDB := 0

	return config.NewRuntimeConfig(grpcAddr, redisAddr, redisPassword, redisDB)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
