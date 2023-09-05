package stream

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/wmattei/go-snake/constants"
)

var cmd *exec.Cmd
var stdinPipe io.WriteCloser

func init() {
	if !constants.SHOULD_STREAM_FRAME {
		return
	}
	rtmpURL := "rtmp://localhost:1935/live/go-snake"

	ffmpegCommand := fmt.Sprintf(
		"ffmpeg -hide_banner -loglevel error -f rawvideo -pixel_format rgb24 -video_size %dx%d -framerate %d -i pipe:0 -c:v libx264 -preset ultrafast -tune zerolatency -f flv %s",
		constants.FRAME_WIDTH,
		constants.FRAME_HEIGHT,
		constants.FPS,
		rtmpURL,
	)

	cmd = exec.Command("bash", "-c", ffmpegCommand)

	_stdinPipe, err := cmd.StdinPipe()
	stdinPipe = _stdinPipe

	if err != nil {
		fmt.Println("Error creating stdin pipe:", err)
		return
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting FFmpeg command:", err)
		return
	}

}

func StreamFrame(frameChannel chan []byte, closeSignal chan bool) {
	if !constants.SHOULD_STREAM_FRAME {
		return
	}

	for {
		select {
		case <-closeSignal:
			err := cmd.Wait()
			if err != nil {
				fmt.Println("Error waiting for FFmpeg command:", err)
			}
			return
		case frame := <-frameChannel:
			_, err := stdinPipe.Write(frame)
			if err != nil {
				fmt.Println("Error writing to pipe:", err)
				break
			}
		}
	}
}
