package stream

import (
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/wmattei/go-snake/constants"
)

func StartStreaming(encodedFrameCh chan []byte, videoTrack *webrtc.TrackLocalStaticSample) {
	for {
		encodedFrame := <-encodedFrameCh
		if encodedFrame == nil {
			break
		}
		videoTrack.WriteSample(media.Sample{Data: encodedFrame, Duration: time.Second / constants.FPS, Timestamp: time.Now()})
	}
}
