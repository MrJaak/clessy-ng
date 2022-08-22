package utils

import (
	"log"
	"os"
)

func RequireEnv(name string) string {
	val, ok := os.LookupEnv(name)
	if !ok {
		log.Fatalf("FATAL: Required env var %s is missing", name)
	}
	if val == "" {
		log.Fatalf("FATAL: Required env var %s is empty", name)
	}
	return val
}

func EnvFallback(name string, fallback string) string {
	val := os.Getenv(name)
	if val == "" {
		return fallback
	}
	return val
}
