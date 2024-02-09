package main

import (
	"log"
	"net/http"
	"simple-chatroom/internal/handlers"
)

func main() {

	log.Println("Starting channel listener")
	go handlers.ListenToWsChannel()

	log.Println("Starting web serve on port 8080")
	err := http.ListenAndServe(":8080", routes())
	if err != nil {
		log.Println(err)
	}
}
