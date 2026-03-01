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

func main() {
	port := config.MustEnv("PORT")
	botToken := config.MustEnv("BOT_TOKEN")
	webhookPath := config.MustEnv("WEBHOOK_PATH")
	storageID := config.MustEnv("STORAGE_ID")

	ctx := context.Background()
	gcs := storage.NewGCS(ctx, storageID)
	client := repository.NewClient(ctx)
	defer client.Close()

	tgBot := bot.NewBot(botToken, client, gcs, ctx)

	// Запускаем worker для обработки очереди обновлений
	tgBot.StartWorkers(1, 100)

	// Регистрируем HTTP handler
	http.HandleFunc(webhookPath, tgBot.WebhookHandler())

	go server.StartHTTPServer(port)    // старт сервера (сразу)
	go tgBot.SetupWebhook(webhookPath) // регистрация вебхука (с задержкой)

	select {}
}
