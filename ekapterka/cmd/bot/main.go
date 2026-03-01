// Package main содержит entrypoint production/runtime-приложения Telegram-бота.
// Файл собирает зависимости (конфиг, Firestore, GCS, bot), поднимает webhook endpoint
// и запускает обработку входящих обновлений.
package main

import (
	"net/http"

	"context"

	"ekapterka/internal/bot"
	"ekapterka/internal/config"
	"ekapterka/internal/repository"
	"ekapterka/internal/server"

	"ekapterka/internal/storage"
)

// main запускает Telegram-бота в webhook-режиме.
// Здесь собираются все зависимости приложения и поднимается HTTP endpoint,
// на который Telegram отправляет обновления.
func main() {
	// Обязательные runtime-настройки.
	port := config.MustEnv("PORT")
	botToken := config.MustEnv("BOT_TOKEN")
	webhookPath := config.MustEnv("WEBHOOK_PATH")
	storageID := config.MustEnv("STORAGE_ID")
	firestoreProjectID := config.MustEnv("FIRESTORE_PROJECT_ID")

	ctx := context.Background()

	// Инициализируем внешние клиенты (GCS + Firestore).
	gcs := storage.NewGCS(ctx, storageID)
	client := repository.NewClient(ctx, firestoreProjectID)
	defer client.Close()

	// Собираем объект бота с зависимостями.
	tgBot := bot.NewBot(botToken, client, gcs, ctx)

	// Запускаем worker'ы для обработки входящих обновлений из очереди.
	tgBot.StartWorkers(1, 100)

	// Регистрируем webhook-обработчик по заданному пути.
	http.HandleFunc(webhookPath, tgBot.WebhookHandler())

	go server.StartHTTPServer(port)    // Старт HTTP сервера.
	go tgBot.SetupWebhook(webhookPath) // Регистрация webhook в Telegram API.

	// Блокируем main-горутина: сервис работает как long-running процесс.
	select {}
}
