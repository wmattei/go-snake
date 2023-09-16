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
	"github.com/wmattei/go-snake/snake_errors"
	"github.com/wmattei/go-snake/snake_webrtc"
	"github.com/wmattei/go-snake/stream"
)

const (
	port = 4000
)

func main() {
	http.HandleFunc("/ws", handleWebsocketConnection)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	snake_errors.HandleError(err)
}

func handleWebsocketConnection(w http.ResponseWriter, r *http.Request) {
	peerConnection, err := snake_webrtc.CreateAndNegotiatePeerConnection(w, r)
	snake_errors.HandleError(err)

	track := peerConnection.GetSenders()[0].Track().(*webrtc.TrackLocalStaticSample)
	fmt.Println("Peer connection established")

	handleDataChannel(peerConnection, track)
}

func handleDataChannel(peerConnection *webrtc.PeerConnection, track *webrtc.TrackLocalStaticSample) {
	peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		fmt.Println("Data channel established")
		closeSignal := make(chan bool)

		commandChannel := make(chan string)
		gameStateCh := make(chan *game.GameState, 1)
		pixelCh := make(chan []byte)
		encodedFrameCh := make(chan []byte)

		gameLoop := game.NewGameLoop(&game.GameLoopInit{CommandChannel: commandChannel, GameStateChannel: gameStateCh, CloseSignal: closeSignal})
		go gameLoop.Start()

		go renderer.StartFrameRenderer(gameStateCh, pixelCh)
		go encoder.StartEncoder(pixelCh, encodedFrameCh)
		go stream.StartStreaming(encodedFrameCh, track)

		go handleChannelClose(dataChannel, peerConnection, gameStateCh, commandChannel, pixelCh, encodedFrameCh, closeSignal)

		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
			handleDataChannelMessage(msg, commandChannel)
		})
	})
}

func handleChannelClose(dataChannel *webrtc.DataChannel, peerConnection *webrtc.PeerConnection, gameStateCh chan *game.GameState, commandChannel chan string, pixelCh chan []byte, encodedFrameCh chan []byte, closeSignal chan bool) {
	<-closeSignal
	fmt.Println("Closing peer connection")
	dataChannel.Close()
	peerConnection.Close()

	close(gameStateCh)
	close(commandChannel)

	// Wait for a second for remaining encoded frames to be sent
	time.Sleep(1 * time.Second)
	close(pixelCh)
	close(encodedFrameCh)
}

func handleDataChannelMessage(msg webrtc.DataChannelMessage, commandChannel chan string) {
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
}
