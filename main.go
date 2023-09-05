package main

import (
	"log"
	"net/http"

	"github.com/wmattei/go-snake/communication"
	"github.com/wmattei/go-snake/game"
	"github.com/wmattei/go-snake/stream"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	closeSignal := make(chan bool)

	comm := communication.WebRTCCommunication{}
	err := comm.Initialize()
	handleError(err)

	comm.Connect()

	commandsMessagesCh := make(chan communication.Message)
	comm.Listen(commandsMessagesCh, "commandsChannel")

	// Convert msg channel to string channel
	commandChannel := make(chan string)
	go func() {
		for msg := range commandsMessagesCh {
			commandChannel <- msg.Data.(string)
		}
		close(commandChannel)
	}()

	frameChannel := make(chan []byte)
	go game.StartGameLoop(frameChannel, commandChannel, closeSignal)
	go stream.StreamFrame(frameChannel, closeSignal)

	http.ListenAndServe(":4000", nil)
}
