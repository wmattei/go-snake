package snake_webrtc

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/wmattei/go-snake/snake_errors"
)

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // TODO: Implement security measures
}

func CreateAndNegotiatePeerConnection(w http.ResponseWriter, r *http.Request) (*webrtc.PeerConnection, error) {
	connectionEstablished := make(chan bool)
	var peerConnection *webrtc.PeerConnection

	var jsonWriterMutex sync.Mutex
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	writeWssMessage := func(messageType string, data interface{}) {
		jsonWriterMutex.Lock()
		defer jsonWriterMutex.Unlock()
		conn.WriteJSON(Message{Type: messageType, Data: data})
	}

	var pc *webrtc.PeerConnection
	go func() {
		for {
			var msg Message
			if err := conn.ReadJSON(&msg); err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) || websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
					break
				} else {
					return
				}
			}
			if msg.Type == "" {
				continue
			}
			if msg.Type == "offer" {
				config := webrtc.Configuration{
					ICEServers: []webrtc.ICEServer{
						{
							URLs: []string{"stun:stun.l.google.com:19302"},
						},
					},
				}

				pc, err = webrtc.NewPeerConnection(config)
				if err != nil {
					writeWssMessage("error", "Error creating peer connection")
					return
				}

				videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "game")
				if err != nil {
					snake_errors.HandleError(err)
					return
				}

				_, err = pc.AddTrack(videoTrack)
				if err != nil {
					snake_errors.HandleError(err)
					return
				}

				pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
					if state == webrtc.PeerConnectionStateConnected {
						peerConnection = pc
						connectionEstablished <- true
					}
				})

				offer := msg.Data.(string)
				pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: offer})

				answer, err := pc.CreateAnswer(nil)
				if err != nil {
					snake_errors.HandleError(err)
					return
				}

				err = pc.SetLocalDescription(answer)
				if err != nil {
					snake_errors.HandleError(err)
					return
				}

				writeWssMessage("answer", answer.SDP)
			}

			if msg.Type == "ice" {
				ice := msg.Data.(string)
				pc.AddICECandidate(webrtc.ICECandidateInit{Candidate: ice})
			}
		}
	}()

	<-connectionEstablished
	return peerConnection, nil
}
