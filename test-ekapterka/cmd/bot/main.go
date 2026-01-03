package main

import (
	"net/http"

	"test-ekapterka/internal/bot"
	"test-ekapterka/internal/config"
	"test-ekapterka/internal/server"
)

func main() {
	port := config.MustEnv("PORT")
	botToken := config.MustEnv("BOT_TOKEN")
	webhookPath := config.MustEnv("WEBHOOK_PATH")

	tgBot := bot.NewBot(botToken)

	// Запускаем worker для обработки очереди обновлений
	tgBot.StartWorkers(1, 100)

	// Регистрируем HTTP handler
	http.HandleFunc(webhookPath, tgBot.WebhookHandler())

	go server.StartHTTPServer(port)    // старт сервера (сразу)
	go tgBot.SetupWebhook(webhookPath) // регистрация вебхука (с задержкой)

	select {}
}
