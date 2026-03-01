package bot

// Файл описывает очередь update-ов и worker-модель обработки:
// прием в buffered channel и последовательный consume воркерами.

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var updateQueue chan *tgbotapi.Update

// initQueue лениво инициализирует внутреннюю очередь обновлений.
func initQueue(size int) {
	if updateQueue == nil {
		updateQueue = make(chan *tgbotapi.Update, size)
	}
}

// EnqueueUpdate кладет update в очередь.
// При переполнении update отбрасывается (drop) и фиксируется в логах.
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

// StartWorkers запускает numWorkers горутин, которые читают очередь
// и последовательно обрабатывают полученные update.
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
