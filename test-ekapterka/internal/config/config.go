package config

import (
	"os"
)

func MustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(key + " is not set")
	}
	return val
}
