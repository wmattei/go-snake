package encodingutil

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/wmattei/go-snake/constants"
	"github.com/wmattei/go-snake/shared/debugutil"
	"github.com/wmattei/go-snake/shared/gameutil"
	"github.com/wmattei/go-snake/shared/logutil"
	"github.com/wmattei/go-snake/shared/webrtcutil"
)

type Canvas struct {
	Data      []byte
	Timestamp time.Time
}

type Encoder struct {
	encodedFrameCh chan *webrtcutil.Streamable
	canvasCh       <-chan *Canvas
	closeSignal    <-chan bool
	gameMetadata   *gameutil.GameMetadata
	debugger       *debugutil.Debugger
}

type EncoderOptions struct {
	EncodedFrameChannel chan *webrtcutil.Streamable
	CanvasChannel       <-chan *Canvas
	CloseSignal         <-chan bool
	GameMetadata        *gameutil.GameMetadata
	Debugger            *debugutil.Debugger
}

func NewEncoder(options *EncoderOptions) *Encoder {
	return &Encoder{
		encodedFrameCh: options.EncodedFrameChannel,
		canvasCh:       options.CanvasChannel,
		gameMetadata:   options.GameMetadata,
		closeSignal:    options.CloseSignal,
		debugger:       options.Debugger,
	}
}

const ffmpegBaseCommand = "ffmpeg -hide_banner -loglevel error -re -f rawvideo -pixel_format rgb24 -video_size %dx%d -framerate %v -r %v -i pipe:0 -c:v libx264 -preset ultrafast -tune zerolatency -bufsize 1000k -g 20 -keyint_min 10 -f h264 pipe:1"

func (e *Encoder) Start() {
	ffmpegCommand := fmt.Sprintf(ffmpegBaseCommand, e.gameMetadata.WindowWidth, e.gameMetadata.WindowHeight, constants.FPS, constants.FPS)
	cmd := exec.Command("bash", "-c", ffmpegCommand)
	cmd.Stderr = os.Stderr
	inPipe, err := cmd.StdinPipe()
	logutil.LogFatal(err)
	outPipe, err := cmd.StdoutPipe()
	logutil.LogFatal(err)
	err = cmd.Start()
	logutil.LogFatal(err)

	go e.writeToFFmpeg(inPipe)
	go e.streamToWebRTCTrack(outPipe)
}

func (e *Encoder) writeToFFmpeg(inPipe io.WriteCloser) {
	for canvas := range e.canvasCh {
		select {
		case <-e.encodedFrameCh: // Check if there's a backlog.
			e.debugger.ReportDroppedFrame()
			continue
		default:
			_, err := inPipe.Write(canvas.Data)
			logutil.LogFatal(err)
		}
	}

	inPipe.Close()
}

func (e *Encoder) streamToWebRTCTrack(outPipe io.Reader) {
	buf := make([]byte, 1024*8)
	for {
		timestamp := time.Now()
		n, err := outPipe.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			logutil.LogFatal(fmt.Errorf("error reading from FFmpeg: %v", err))
			continue
		}

		e.encodedFrameCh <- &webrtcutil.Streamable{Data: buf[:n], Timestamp: timestamp}
	}
}
