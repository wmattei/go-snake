package signaling_server

import (
	"github.com/pion/webrtc/v3"
)

type SignalingServer interface {
	Start() error
	OnOfferReceived(offer webrtc.SessionDescription) error
	SendAnswer() error
	SendIceCandidate(candidate webrtc.ICECandidateInit) error
	OnIceCandidateReceived(candidate webrtc.ICECandidateInit) error
	OnDataChannelEstablished(callback func(dataChannel *webrtc.DataChannel))
	GetVideoTrack() *webrtc.TrackLocalStaticSample
	Close() error
}
