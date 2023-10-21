package webrtcutil

import (
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/wmattei/go-snake/constants"
	"github.com/wmattei/go-snake/shared/debugutil"
)

type Streamable struct {
	Data      []byte
	Timestamp time.Time
}

func streamFrame(encodedFrame *Streamable, videoTrack *webrtc.TrackLocalStaticSample) {
	// started := time.Now()
	// defer logutil.LogTimeElapsed(started, "Frame streaming took: ")
	videoTrack.WriteSample(media.Sample{Data: encodedFrame.Data, Duration: time.Second / constants.FPS, Timestamp: encodedFrame.Timestamp})
}

func StartStreaming(encodedFrameCh chan *Streamable, videoTrack *webrtc.TrackLocalStaticSample, debugger *debugutil.Debugger) {
	var latestTimestamp time.Time

	go func() {
		for {
			encodedFrame := <-encodedFrameCh
			if encodedFrame == nil {
				break
			}

			if encodedFrame.Timestamp.Before(latestTimestamp) {
				// Drop the frame
				// fmt.Println("Dropping frame..")
				continue
			}

			debugger.ReportFrameStream()
			latestTimestamp = encodedFrame.Timestamp
			streamFrame(encodedFrame, videoTrack)
		}
	}()
}
