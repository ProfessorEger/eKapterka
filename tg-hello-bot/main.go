package main

import (
	"log"
	"net/http"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// ---- PORT (Cloud Run requirement)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // локально
	}

	// ---- ENV
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN is not set")
	}

	webhookPath := os.Getenv("WEBHOOK_PATH")
	if webhookPath == "" {
		log.Fatal("WEBHOOK_PATH is not set")
	}

	// ---- Telegram bot
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	// ---- HTTP handler (Telegram webhook)
	http.HandleFunc(webhookPath, telegramWebhookHandler(bot))

	// ---- START SERVER (ОБЯЗАТЕЛЬНО СРАЗУ)
	go startHTTPServer(port)

	// ---- SET WEBHOOK (после старта сервера)
	go setupWebhook(bot, webhookPath)

	select {}
}

func telegramWebhookHandler(bot *tgbotapi.BotAPI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		update, err := bot.HandleUpdate(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello, World!")
			bot.Send(msg)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func startHTTPServer(port string) {
	log.Println("Listening on :" + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func setupWebhook(bot *tgbotapi.BotAPI, webhookPath string) {
	time.Sleep(5 * time.Second)

	serviceURL := os.Getenv("SERVICE_URL")
	if serviceURL == "" {
		log.Println("SERVICE_URL not set, skipping setWebhook")
		return
	}

	webhookURL := serviceURL + webhookPath
	log.Println("Setting webhook to:", webhookURL)

	wh, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		log.Println("Webhook config error:", err)
		return
	}

	_, err = bot.Request(wh)
	if err != nil {
		log.Println("setWebhook failed:", err)
		return
	}

	log.Println("Webhook set successfully")
}
