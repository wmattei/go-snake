package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pion/webrtc/v3"
	latencycheck "github.com/wmattei/go-snake/games/latency_check"
	"github.com/wmattei/go-snake/shared/debugutil"
	"github.com/wmattei/go-snake/shared/encodingutil"
	"github.com/wmattei/go-snake/shared/logutil"
	"github.com/wmattei/go-snake/shared/webrtcutil"
)

const (
	port = 4000
)

func main() {
	http.HandleFunc("/ws", handleWebsocketConnection)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	logutil.LogFatal(err)
}

func handleWebsocketConnection(w http.ResponseWriter, r *http.Request) {
	peerConnection, err := webrtcutil.CreateAndNegotiatePeerConnection(w, r)
	logutil.LogFatal(err)

	track := peerConnection.GetSenders()[0].Track().(*webrtc.TrackLocalStaticSample)
	fmt.Println("Peer connection established")

	handleDataChannel(peerConnection, track)
}

func handleDataChannel(peerConnection *webrtc.PeerConnection, track *webrtc.TrackLocalStaticSample) {
	peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		fmt.Println("Data channel established")
		closeSignal := make(chan bool)

		commandChannel := make(chan interface{})
		canvasCh := make(chan *encodingutil.Canvas)
		encodedFrameCh := make(chan *webrtcutil.Streamable)

		go handleChannelClose(dataChannel, peerConnection, commandChannel, canvasCh, encodedFrameCh, closeSignal)
		debugger := debugutil.NewDebugger(500)
		go debugger.StartDebugger()

		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
			var message webrtcutil.Message
			err := json.Unmarshal(msg.Data, &message)
			if err != nil {
				fmt.Println("Error unmarshalling message:", err)
				return
			}

			if message.Type == "start" {
				windowWidth := int(message.Data.(map[string]interface{})["width"].(float64))
				windowHeight := int(message.Data.(map[string]interface{})["height"].(float64))
				go latencycheck.StartLatencyCheck(&latencycheck.LatencyCheckInit{
					WindowWidth:    windowWidth,
					WindowHeight:   windowHeight,
					CommandChannel: commandChannel,
					CanvasChannel:  canvasCh,
					CloseSignal:    closeSignal,
				})
				go encodingutil.StartEncoder(canvasCh, encodedFrameCh, windowWidth, windowHeight)
				go webrtcutil.StartStreaming(encodedFrameCh, track, debugger)
			} else {
				// fmt.Println(message.Data.(map[string]interface{})["position"])
				commandChannel <- message.Data
			}
		})
	})
}

func handleChannelClose(dataChannel *webrtc.DataChannel, peerConnection *webrtc.PeerConnection, commandChannel chan interface{}, canvasCh chan *encodingutil.Canvas, encodedFrameCh chan *webrtcutil.Streamable, closeSignal chan bool) {
	<-closeSignal
	fmt.Println("Closing peer connection")
	dataChannel.Close()
	peerConnection.Close()

	close(commandChannel)

	// Wait for a second for remaining encoded frames to be sent
	time.Sleep(1 * time.Second)
	close(canvasCh)
	close(encodedFrameCh)
}
