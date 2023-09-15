package stream

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264reader"
	"github.com/wmattei/go-snake/constants"
	"github.com/wmattei/go-snake/snake_errors"
)

// var cmd *exec.Cmd
// var stdinPipe io.WriteCloser

// func init() {
// 	if !constants.SHOULD_STREAM_FRAME {
// 		return
// 	}
// 	rtmpURL := "rtmp://localhost:1935/live/go-snake"

// 	ffmpegCommand := fmt.Sprintf(
// 		"ffmpeg -hide_banner -loglevel error -f rawvideo -pixel_format rgb24 -video_size %dx%d -framerate %d -i pipe:0 -c:v libx264 -preset ultrafast -tune zerolatency -f flv %s",
// 		constants.FRAME_WIDTH,
// 		constants.FRAME_HEIGHT,
// 		constants.FPS,
// 		rtmpURL,
// 	)

// 	cmd = exec.Command("bash", "-c", ffmpegCommand)

// 	_stdinPipe, err := cmd.StdinPipe()
// 	stdinPipe = _stdinPipe

// 	if err != nil {
// 		fmt.Println("Error creating stdin pipe:", err)
// 		return
// 	}

// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	if err := cmd.Start(); err != nil {
// 		fmt.Println("Error starting FFmpeg command:", err)
// 		return
// 	}

// }

func StreamFrame(frameChannel chan []byte, closeSignal chan bool, videoTrack *webrtc.TrackLocalStaticSample) {
	if !constants.SHOULD_STREAM_FRAME {
		return
	}

	for {
		select {
		case <-closeSignal:
			return
		case frame := <-frameChannel:
			const ffmpegCommand = "ffmpeg -hide_banner -loglevel error -f rawvideo -pixel_format rgb24 -video_size %dx%d -framerate %d -i pipe:0 -c:v libx264 -preset ultrafast -tune zerolatency -f h264 pipe:1"
			cmd := exec.Command("bash", "-c", fmt.Sprintf(ffmpegCommand, constants.FRAME_WIDTH, constants.FRAME_HEIGHT, constants.FPS))

			cmd.Stderr = os.Stderr

			inPipe, err := cmd.StdinPipe()
			snake_errors.HandleError(err)

			outPipe, err := cmd.StdoutPipe()
			snake_errors.HandleError(err)

			err = cmd.Start()
			snake_errors.HandleError(err)

			_, err = inPipe.Write(frame)
			snake_errors.HandleError(err)

			inPipe.Close()

			h264, err := h264reader.NewReader(outPipe)
			snake_errors.HandleError(err)

			data := []byte{}
			spsAndPpsCache := []byte{}
			for {
				nal, h264Err := h264.NextNAL()
				if h264Err == io.EOF {
					// Finished sending frames
					break
				} else if h264Err != nil {
					fmt.Printf("Error reading H.264 NAL: %v\n", h264Err)
					break
				}

				nal.Data = append([]byte{0x00, 0x00, 0x00, 0x01}, nal.Data...)
				if nal.UnitType == h264reader.NalUnitTypeSPS || nal.UnitType == h264reader.NalUnitTypePPS {
					spsAndPpsCache = append(spsAndPpsCache, nal.Data...)
					continue
				} else if nal.UnitType == h264reader.NalUnitTypeCodedSliceIdr {
					nal.Data = append(spsAndPpsCache, nal.Data...)
					spsAndPpsCache = []byte{}
				}

				data = append(data, nal.Data...)
			}

			videoTrack.WriteSample(media.Sample{Data: data, Duration: time.Second / constants.FPS})
			cmd.Wait()
		}
	}
}
