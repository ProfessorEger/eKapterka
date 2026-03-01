// Package server содержит HTTP-инфраструктуру приложения.
// В текущем проекте используется для приема Telegram webhook-запросов.
package server

// Файл отвечает за запуск HTTP-сервера процесса.

import (
	"log"
	"net/http"
)

// StartHTTPServer поднимает HTTP сервер для webhook-обработчика Telegram.
// Конкретные handler'ы регистрируются в cmd/bot/main.go через http.HandleFunc.
func StartHTTPServer(port string) {
	log.Println("Listening on :" + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
