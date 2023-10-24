package encodingutil

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/wmattei/go-snake/constants"
	"github.com/wmattei/go-snake/shared/debugutil"
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
	windowWidth    int
	windowHeight   int
	debugger       *debugutil.Debugger
}

type EncoderOptions struct {
	EncodedFrameChannel chan *webrtcutil.Streamable
	CanvasChannel       <-chan *Canvas
	CloseSignal         <-chan bool
	WindowWidth         int
	WindowHeight        int
	Debugger            *debugutil.Debugger
}

func NewEncoder(options *EncoderOptions) *Encoder {
	return &Encoder{
		encodedFrameCh: options.EncodedFrameChannel,
		canvasCh:       options.CanvasChannel,
		closeSignal:    options.CloseSignal,
		debugger:       options.Debugger,
		windowWidth:    options.WindowWidth,
		windowHeight:   options.WindowHeight,
	}
}

const ffmpegBaseCommand = "ffmpeg -hide_banner -loglevel error -re -f rawvideo -pixel_format rgb24 -video_size %dx%d -framerate %v -r %v -i pipe:0 -c:v libx264 -preset ultrafast -tune zerolatency -bufsize 1000k -g 10 -keyint_min 5 -maxrate 1000k -f h264 pipe:1"

func (e *Encoder) Start() {
	ffmpegCommand := fmt.Sprintf(ffmpegBaseCommand, e.windowWidth, e.windowHeight, constants.FPS, constants.FPS)
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
