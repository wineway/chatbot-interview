package main

import (
	"log"
	"net/http"
	"os"

	"github.com/wineway/chatbot/pkg/messenger_service"
)

func main() {
	port := os.Getenv("CHATBOT_PORT")
	token := os.Getenv("CHATBOT_TOKEN")
	accessToken := os.Getenv("CHATBOT_ACCESS_TOKEN")
	s := messenger_service.NewMessengerService(token, accessToken)
	log.Fatal(http.ListenAndServe("localhost:"+port, s.Handler()))
}
