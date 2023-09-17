package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/wmattei/go-snake/games/snake"
	"github.com/wmattei/go-snake/shared/encodingutil"
	"github.com/wmattei/go-snake/shared/errutil"
	"github.com/wmattei/go-snake/shared/webrtcutil"
)

const (
	port = 4000
)

func main() {
	http.HandleFunc("/ws", handleWebsocketConnection)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	errutil.HandleError(err)
}

func handleWebsocketConnection(w http.ResponseWriter, r *http.Request) {
	peerConnection, err := webrtcutil.CreateAndNegotiatePeerConnection(w, r)
	errutil.HandleError(err)

	track := peerConnection.GetSenders()[0].Track().(*webrtc.TrackLocalStaticSample)
	fmt.Println("Peer connection established")

	handleDataChannel(peerConnection, track)
}

func handleDataChannel(peerConnection *webrtc.PeerConnection, track *webrtc.TrackLocalStaticSample) {
	peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		fmt.Println("Data channel established")
		closeSignal := make(chan bool)

		commandChannel := make(chan string)
		frameChannel := make(chan []byte)
		encodedFrameCh := make(chan []byte)

		go snake.StartSnakeGame(commandChannel, frameChannel, closeSignal)

		go encodingutil.StartEncoder(frameChannel, encodedFrameCh)
		go webrtcutil.StartStreaming(encodedFrameCh, track)

		go handleChannelClose(dataChannel, peerConnection, commandChannel, frameChannel, encodedFrameCh, closeSignal)

		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
			handleDataChannelMessage(msg, commandChannel)
		})
	})
}

func handleChannelClose(dataChannel *webrtc.DataChannel, peerConnection *webrtc.PeerConnection, commandChannel chan string, pixelCh chan []byte, encodedFrameCh chan []byte, closeSignal chan bool) {
	<-closeSignal
	fmt.Println("Closing peer connection")
	dataChannel.Close()
	peerConnection.Close()

	close(commandChannel)

	// Wait for a second for remaining encoded frames to be sent
	time.Sleep(1 * time.Second)
	close(pixelCh)
	close(encodedFrameCh)
}

func handleDataChannelMessage(msg webrtc.DataChannelMessage, commandChannel chan string) {
	var message webrtcutil.Message
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
