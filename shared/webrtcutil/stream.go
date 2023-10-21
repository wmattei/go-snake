package webrtcutil

import (
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/wmattei/go-snake/shared/debugutil"
)

func streamFrame(encodedFrame []byte, videoTrack *webrtc.TrackLocalStaticSample) {
	// started := time.Now()
	// defer logutil.LogTimeElapsed(started, "Frame streaming took: ")
	videoTrack.WriteSample(media.Sample{Data: encodedFrame, Duration: time.Millisecond, Timestamp: time.Now()})
}

func StartStreaming(encodedFrameCh chan []byte, videoTrack *webrtc.TrackLocalStaticSample, debugger *debugutil.Debugger) {
	for {
		encodedFrame := <-encodedFrameCh
		if encodedFrame == nil {
			break
		}
		debugger.ReportFrameStream()
		streamFrame(encodedFrame, videoTrack)
	}
}
