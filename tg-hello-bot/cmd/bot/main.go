package main

import (
	"net/http"

	"tg-hello-bot/internal/config"
	"tg-hello-bot/internal/server"
	"tg-hello-bot/internal/telegram"
)

func main() {
	port := config.MustEnv("PORT")
	botToken := config.MustEnv("BOT_TOKEN")
	webhookPath := config.MustEnv("WEBHOOK_PATH")

	bot := telegram.NewTelegramBot(botToken)

	// Регистрируем HTTP handler
	http.HandleFunc(webhookPath, telegram.WebhookHandler(bot))

	go server.StartHTTPServer(port)            // старт сервера (сразу)
	go telegram.SetupWebhook(bot, webhookPath) // регистрация вебхука (с задержкой)

	select {}
}
