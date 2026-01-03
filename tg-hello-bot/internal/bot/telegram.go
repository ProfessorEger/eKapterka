package bot

import (
	"log"
	"net/http"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SetupWebhook(bot *tgbotapi.BotAPI, webhookPath string) {
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

func WebhookHandler(bot *tgbotapi.BotAPI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		update, err := bot.HandleUpdate(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		EnqueueUpdate(update)

		w.WriteHeader(http.StatusOK)
	}
}

func NewTelegramBot(token string) *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}
	return bot
}
