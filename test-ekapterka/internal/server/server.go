package server

import (
	"log"
	"net/http"
)

func StartHTTPServer(port string) {
	log.Println("Listening on :" + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
