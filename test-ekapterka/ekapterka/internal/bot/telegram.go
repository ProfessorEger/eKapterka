package bot

import (
	"ekapterka/internal/repository"
	"log"
	"net/http"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api  *tgbotapi.BotAPI
	repo *repository.Client // или конкретный firestore.Client
}

func NewBot(token string, client *repository.Client) *Bot {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	return &Bot{
		api:  api,
		repo: client,
	}
}

func (b *Bot) SetupWebhook(webhookPath string) {
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

	_, err = b.api.Request(wh)
	if err != nil {
		log.Println("setWebhook failed:", err)
		return
	}

	log.Println("Webhook set successfully")
}

func (b *Bot) WebhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		update, err := b.api.HandleUpdate(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		EnqueueUpdate(update)

		w.WriteHeader(http.StatusOK)
	}
}
