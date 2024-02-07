package signaling_server

import "github.com/pion/webrtc/v3"

type SignalingServer interface {
	Connect(callback func()) error
	GetVideoTrack() *webrtc.TrackLocalStaticSample
	GetDataChannel() *webrtc.DataChannel
}
