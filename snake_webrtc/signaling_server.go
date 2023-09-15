package snake_webrtc

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264reader"
	"github.com/wmattei/go-snake/snake_errors"
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

func handleTrack(pc *webrtc.PeerConnection) {
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "game")
	snake_errors.HandleError(err)
	rtpSender, err := pc.AddTrack(videoTrack)
	snake_errors.HandleError(err)

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	go func() {
		cmd := exec.Command("ffmpeg",
			"-i", "./earth.mp4",
			"-f", "h264",
			"pipe:1")
		dataPipe, err := cmd.StdoutPipe()
		snake_errors.HandleError(err)
		err = cmd.Start()
		snake_errors.HandleError(err)

		h264, err := h264reader.NewReader(dataPipe)
		snake_errors.HandleError(err)

		spsAndPpsCache := []byte{}
		ticker := time.NewTicker(time.Millisecond * 10)
		for ; true; <-ticker.C {
			nal, h264Err := h264.NextNAL()
			if h264Err == io.EOF {
				fmt.Printf("All video frames parsed and sent")
				break
			}
			if h264Err != nil {
				panic(h264Err)
			}

			nal.Data = append([]byte{0x00, 0x00, 0x00, 0x01}, nal.Data...)

			if nal.UnitType == h264reader.NalUnitTypeSPS || nal.UnitType == h264reader.NalUnitTypePPS {
				spsAndPpsCache = append(spsAndPpsCache, nal.Data...)
				continue
			} else if nal.UnitType == h264reader.NalUnitTypeCodedSliceIdr {
				nal.Data = append(spsAndPpsCache, nal.Data...)
				spsAndPpsCache = []byte{}
			}
			if h264Err = videoTrack.WriteSample(media.Sample{Data: nal.Data, Duration: time.Second}); h264Err != nil {
				panic(h264Err)
			}
		}
	}()
}

func HandleWebRtcSignaling(onPeerConnection func(*webrtc.PeerConnection, *webrtc.TrackLocalStaticSample)) {
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
				config := webrtc.Configuration{
					ICEServers: []webrtc.ICEServer{
						{
							URLs: []string{"stun:stun.l.google.com:19302"},
						},
					},
				}

				peerConnection, err = webrtc.NewPeerConnection(config)
				snake_errors.HandleError(err)

				// handleTrack(peerConnection)

				videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "game")
				snake_errors.HandleError(err)

				rtpSender, err := peerConnection.AddTrack(videoTrack)
				snake_errors.HandleError(err)
				go func() {
					rtcpBuf := make([]byte, 1500)
					for {
						if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
							return
						}
					}
				}()
				peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
					if state == webrtc.PeerConnectionStateConnected {
						onPeerConnection(peerConnection, videoTrack)
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
