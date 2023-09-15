package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pion/webrtc/v3"

	"github.com/wmattei/go-snake/encoder"
	"github.com/wmattei/go-snake/game"
	"github.com/wmattei/go-snake/renderer"
	"github.com/wmattei/go-snake/snake_webrtc"
	"github.com/wmattei/go-snake/stream"
)

func main() {

	snake_webrtc.HandleWebRtcSignaling(func(pc *webrtc.PeerConnection, track *webrtc.TrackLocalStaticSample) {
		fmt.Println("Peer connection established")

		pc.OnDataChannel(func(dc *webrtc.DataChannel) {
			fmt.Println("Data channel established")
			closeSignal := make(chan bool)

			commandChannel := make(chan string)
			gameStateCh := make(chan *game.GameState)
			pixelCh := make(chan []byte)
			encodedFrameCh := make(chan []byte)

			// Is this the best approach for multi-tasking? COULD BE LOL
			go game.StartGameLoop(commandChannel, gameStateCh, closeSignal)
			go renderer.StartFrameRenderer(gameStateCh, pixelCh)
			go encoder.StartEncoder(pixelCh, encodedFrameCh)
			go stream.StartStreaming(encodedFrameCh, track)

			go func() {
				<-closeSignal
				fmt.Println("Closing peer connection")
				close(gameStateCh)
				close(commandChannel)
				close(pixelCh)

				time.Sleep(1 * time.Second)
				close(encodedFrameCh)
				dc.Close()
				pc.Close()
			}()

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
