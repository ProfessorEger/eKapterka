// Package config содержит утилиты чтения конфигурации приложения
// из переменных окружения.
package config

// Файл включает минимальный helper для fail-fast валидации обязательных env.

import (
	"os"
)

// MustEnv читает переменную окружения и паникует, если она не задана.
// Такой подход упрощает fail-fast поведение на старте приложения.
func MustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(key + " is not set")
	}
	return val
}
