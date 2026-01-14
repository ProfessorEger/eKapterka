package main

import (
	"net/http"

	"ekapterka/internal/bot"
	"ekapterka/internal/config"
	"ekapterka/internal/repository"
	"ekapterka/internal/server"

	"context"
)

func main() {
	port := config.MustEnv("PORT")
	botToken := config.MustEnv("BOT_TOKEN")
	webhookPath := config.MustEnv("WEBHOOK_PATH")

	ctx := context.Background()
	client := repository.NewClient(ctx)
	defer client.Close()

	tgBot := bot.NewBot(botToken, client, ctx)

	// Запускаем worker для обработки очереди обновлений
	tgBot.StartWorkers(1, 100)

	// Регистрируем HTTP handler
	http.HandleFunc(webhookPath, tgBot.WebhookHandler())

	go server.StartHTTPServer(port)    // старт сервера (сразу)
	go tgBot.SetupWebhook(webhookPath) // регистрация вебхука (с задержкой)

	select {}
}
