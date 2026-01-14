package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var updateQueue chan *tgbotapi.Update

// initQueue ensures the queue is initialized with given size.
func initQueue(size int) {
	if updateQueue == nil {
		updateQueue = make(chan *tgbotapi.Update, size)
	}
}

// EnqueueUpdate pushes an update into the queue (drops if full).
func EnqueueUpdate(update *tgbotapi.Update) {
	if update == nil {
		return
	}
	if updateQueue == nil {
		initQueue(100)
	}
	select {
	case updateQueue <- update:
	default:
		log.Println("update queue full, dropping update")
	}
}

// StartWorkers launches numWorkers goroutines processing updates from the queue.
func (b *Bot) StartWorkers(numWorkers int, queueSize int) {
	if b.api == nil || numWorkers <= 0 {
		return
	}
	initQueue(queueSize)
	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			for update := range updateQueue {
				if update == nil {
					continue
				}
				b.handleUpdate(update)
			}
		}(i)
	}
}
