package communication

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // TODO security lol
}

type WebRTCCommunication struct {
	peerConnection *webrtc.PeerConnection
	channels       map[string]*webrtc.DataChannel
}

func (w *WebRTCCommunication) Initialize() error {
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&webrtc.MediaEngine{}))
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		return err
	}

	w.peerConnection = peerConnection
	w.channels = make(map[string]*webrtc.DataChannel)
	return nil
}

func (w *WebRTCCommunication) Connect() {
	var jsonWriterMutex sync.Mutex

	http.HandleFunc("/ws", func(writer http.ResponseWriter, r *http.Request) {

		conn, err := upgrader.Upgrade(writer, r, nil)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writeWssMessage := func(messageType string, data interface{}) {
			jsonWriterMutex.Lock()
			defer jsonWriterMutex.Unlock()
			conn.WriteJSON(Message{Type: messageType, Data: data})
		}

		w.peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
			if candidate != nil {
				writeWssMessage("ice", candidate)
			}
		})

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
				offer := msg.Data.(string)
				w.peerConnection.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: offer})

				answer, err := w.peerConnection.CreateAnswer(nil)
				if err != nil {
					return
				}

				err = w.peerConnection.SetLocalDescription(answer)
				if err != nil {
					return
				}

				writeWssMessage("answer", answer.SDP)
			}
			if msg.Type == "ice" {
				ice := msg.Data.(string)
				w.peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: ice})

				// Close websocket connection. From now on, we rely only on webrtc
				// conn.Close()
			}

		}

	})
}

func (w *WebRTCCommunication) SendMessage(message Message) error {
	return nil
}

func (w *WebRTCCommunication) Listen(listener chan Message, channelName string) error {
	if w.channels[channelName] == nil {
		channel, err := w.peerConnection.CreateDataChannel(channelName, nil)
		if err != nil {
			return err
		}
		w.channels[channelName] = channel
	}

	w.peerConnection.OnDataChannel(func(channel *webrtc.DataChannel) {
		if channel.Label() != channelName {
			return
		}
		channel.OnMessage(func(msg webrtc.DataChannelMessage) {
			var message Message
			if err := json.Unmarshal(msg.Data, &message); err != nil {
				fmt.Println("Error decoding message:", err)
				return
			}

			fmt.Println("Received message:", message)
			listener <- message
		})
	})

	return nil
}
