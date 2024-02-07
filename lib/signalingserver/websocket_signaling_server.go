package signaling_server

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/wmattei/go-snake/lib/logutil"
)

type WebSocketSignalingServer struct {
	Port            string
	MimeType        string
	server          *http.Server
	jsonWriterMutex sync.Mutex
	wsConnection    *websocket.Conn
	peerConnection  *webrtc.PeerConnection
	dataChannel     *webrtc.DataChannel

	onReadyCb func()
}

func NewWebSocketSignalingServer(port string, mimeType string) *WebSocketSignalingServer {
	return &WebSocketSignalingServer{
		Port:     port,
		MimeType: mimeType,
	}
}

func (ws *WebSocketSignalingServer) Connect(cb func()) error {
	ws.onReadyCb = cb
	ws.server = &http.Server{
		Addr:    ":" + ws.Port,
		Handler: http.HandlerFunc(ws.handleWebsocketConnection),
	}

	err := ws.server.ListenAndServe()
	if err != http.ErrServerClosed {
		logutil.LogFatal(err)
	}

	return nil
}

func (ws *WebSocketSignalingServer) GetVideoTrack() *webrtc.TrackLocalStaticSample {
	return ws.peerConnection.GetSenders()[0].Track().(*webrtc.TrackLocalStaticSample)
}

func (ws *WebSocketSignalingServer) GetDataChannel() *webrtc.DataChannel {
	return ws.dataChannel
}

func (ws *WebSocketSignalingServer) Close() error {
	ws.peerConnection.Close()
	ws.wsConnection.Close()
	ws.server.Close()
	return nil
}

func (ws *WebSocketSignalingServer) onOfferReceived(offer webrtc.SessionDescription) error {
	ws.createPeerConnection()
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: ws.MimeType}, "video", "game")
	if err != nil {
		logutil.LogFatal(err)
		return err
	}

	_, err = ws.peerConnection.AddTrack(videoTrack)
	if err != nil {
		logutil.LogFatal(err)
		return err
	}

	ws.peerConnection.SetRemoteDescription(offer)
	ws.sendAnswer()

	return nil
}

func (ws *WebSocketSignalingServer) sendAnswer() error {
	answer, err := ws.peerConnection.CreateAnswer(nil)
	if err != nil {
		logutil.LogFatal(err)
		return err
	}

	err = ws.peerConnection.SetLocalDescription(answer)
	if err != nil {
		logutil.LogFatal(err)
		return err
	}

	ws.writeWssMessage("answer", answer.SDP)
	return nil
}

func (ws *WebSocketSignalingServer) sendIceCandidate(candidate webrtc.ICECandidateInit) error {
	ws.writeWssMessage("ice", candidate.Candidate)
	return nil
}

func (ws *WebSocketSignalingServer) onIceCandidateReceived(candidate webrtc.ICECandidateInit) error {
	err := ws.peerConnection.AddICECandidate(candidate)
	logutil.LogFatal(err)
	return err
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // TODO: Implement security measures
}

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (ws *WebSocketSignalingServer) createPeerConnection() error {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	pc, err := webrtc.NewPeerConnection(config)

	if err != nil {
		ws.writeWssMessage("error", "Error creating peer connection")
		return err
	}

	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			ws.sendIceCandidate(candidate.ToJSON())
		}
	})

	pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		ws.dataChannel = dc
		if ws.onReadyCb != nil {
			ws.onReadyCb()
		}

	})

	ws.peerConnection = pc

	return nil
}

func (ws *WebSocketSignalingServer) writeWssMessage(messageType string, data interface{}) error {
	ws.jsonWriterMutex.Lock()
	defer ws.jsonWriterMutex.Unlock()
	err := ws.wsConnection.WriteJSON(WSMessage{Type: messageType, Data: data})
	return err
}

func (ws *WebSocketSignalingServer) upgradeConnection(w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	logutil.LogFatal(err)
	ws.wsConnection = conn
	return nil
}

func (ws *WebSocketSignalingServer) handleWebsocketConnection(w http.ResponseWriter, r *http.Request) {
	ws.upgradeConnection(w, r)

	go func() {
		defer ws.wsConnection.Close()
		for {
			var msg WSMessage
			if err := ws.wsConnection.ReadJSON(&msg); err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) || websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
					break
				} else {
					return
				}
			}
			switch msg.Type {
			case "offer":
				ws.onOfferReceived(webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: msg.Data.(string)})
			case "ice":
				ws.onIceCandidateReceived(webrtc.ICECandidateInit{Candidate: msg.Data.(string)})
			}
		}
	}()

}
