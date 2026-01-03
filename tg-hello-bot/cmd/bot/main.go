package main

import (
	"net/http"

	"tg-hello-bot/internal/bot"
	"tg-hello-bot/internal/config"
	"tg-hello-bot/internal/server"
)

func main() {
	port := config.MustEnv("PORT")
	botToken := config.MustEnv("BOT_TOKEN")
	webhookPath := config.MustEnv("WEBHOOK_PATH")

	tgBot := bot.NewTelegramBot(botToken)

	// Запускаем worker для обработки очереди обновлений
	bot.StartWorkers(tgBot, 1, 100)

	// Регистрируем HTTP handler
	http.HandleFunc(webhookPath, bot.WebhookHandler(tgBot))

	go server.StartHTTPServer(port)         // старт сервера (сразу)
	go bot.SetupWebhook(tgBot, webhookPath) // регистрация вебхука (с задержкой)

	select {}
}
