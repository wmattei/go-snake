package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	pion "github.com/pion/webrtc/v3"
	"github.com/wmattei/go-snake/game"
	"github.com/wmattei/go-snake/stream"
	"github.com/wmattei/go-snake/webrtc"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	webrtc.HandleWebRtcSignaling(func(pc *pion.PeerConnection) {
		fmt.Println("Peer connection established")
		commandChannel := make(chan string)

		frameChannel := make(chan []byte)
		closeSignal := make(chan bool)
		go game.StartGameLoop(frameChannel, commandChannel, closeSignal)
		go stream.StreamFrame(frameChannel, closeSignal)

		pc.OnDataChannel(func(dc *pion.DataChannel) {
			fmt.Println("Data channel established")
			dc.OnMessage(func(msg pion.DataChannelMessage) {
				var message webrtc.Message
				err := json.Unmarshal(msg.Data, &message)
				if err != nil {
					fmt.Println("Error unmarshalling message:", err)
					return
				}
				if message.Type != "command" {
					fmt.Println("Channel used for wrong message type:", message.Type)
					return
				}

				fmt.Println("Received command:", message.Data.(string))
				commandChannel <- message.Data.(string)

			})
		})

	})
	http.ListenAndServe(":4000", nil)
}
