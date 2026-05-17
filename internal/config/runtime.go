package config

import "runtime"

// RuntimeConfig holds all startup-time configuration.
// All values are fixed at construction; no dynamic mutation permitted.
type RuntimeConfig struct {
	WorkerCount   int
	GRPCAddr      string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func NewRuntimeConfig(grpcAddr, redisAddr, redisPassword string, redisDB int) RuntimeConfig {
	return RuntimeConfig{
		WorkerCount:   runtime.NumCPU() * 2,
		GRPCAddr:      grpcAddr,
		RedisAddr:     redisAddr,
		RedisPassword: redisPassword,
		RedisDB:       redisDB,
	}
}
