package webrtc

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // TODO security lol
}

func HandleWebRtcSignaling(onPeerConnection func(*webrtc.PeerConnection)) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {

		var jsonWriterMutex sync.Mutex
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("WebSocket upgrade error:", err)
			return
		}
		defer conn.Close()

		writeWssMessage := func(messageType string, data interface{}) {
			jsonWriterMutex.Lock()
			defer jsonWriterMutex.Unlock()
			conn.WriteJSON(Message{Type: messageType, Data: data})
		}

		var peerConnection *webrtc.PeerConnection

		for {
			var msg Message
			if err := conn.ReadJSON(&msg); err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					break
				} else if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
					break
				} else {
					return
				}
			}

			if msg.Type == "" {
				continue
			}
			if msg.Type == "offer" {
				api := webrtc.NewAPI(webrtc.WithMediaEngine(&webrtc.MediaEngine{}))
				config := webrtc.Configuration{
					ICEServers: []webrtc.ICEServer{
						{
							URLs: []string{"stun:stun.l.google.com:19302"},
						},
					},
				}

				peerConnection, err = api.NewPeerConnection(config)
				peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
					if state == webrtc.PeerConnectionStateConnected {
						onPeerConnection(peerConnection)
					}
				})
				if err != nil {
					writeWssMessage("error", "Error creating peer connection")
					return
				}

				offer := msg.Data.(string)
				peerConnection.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: offer})

				answer, _ := peerConnection.CreateAnswer(nil)

				peerConnection.SetLocalDescription(answer)

				writeWssMessage("answer", answer.SDP)
			}
			if msg.Type == "ice" {
				ice := msg.Data.(string)
				peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: ice})
			}

		}
	})
}
