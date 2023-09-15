package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pion/webrtc/v3"

	"github.com/wmattei/go-snake/game"
	"github.com/wmattei/go-snake/snake_webrtc"
	"github.com/wmattei/go-snake/stream"
)

func main() {

	snake_webrtc.HandleWebRtcSignaling(func(pc *webrtc.PeerConnection, track *webrtc.TrackLocalStaticSample) {
		fmt.Println("Peer connection established")

		pc.OnDataChannel(func(dc *webrtc.DataChannel) {
			fmt.Println("Data channel established")
			commandChannel := make(chan string)
			frameChannel := make(chan []byte)
			closeSignal := make(chan bool)

			go game.StartGameLoop(frameChannel, commandChannel, closeSignal)
			go stream.StreamFrame(frameChannel, closeSignal, track)

			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				var message snake_webrtc.Message
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
