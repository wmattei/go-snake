package encodingutil

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wmattei/go-snake/constants"
	"github.com/wmattei/go-snake/lib/debugutil"
	"github.com/wmattei/go-snake/lib/logutil"
	"github.com/wmattei/go-snake/lib/webrtcutil"
)

type Canvas struct {
	Data      []byte
	Timestamp time.Time
}

type Dimensions struct {
	Width  int
	Height int
}

type Encoder struct {
	encodedFrameCh chan *webrtcutil.Streamable
	canvasCh       <-chan *Canvas
	closeSignal    <-chan bool
	debugger       *debugutil.Debugger
	closed         int32
	wg             sync.WaitGroup

	dimensionsChannel <-chan *Dimensions
	cmd               *exec.Cmd
}

type EncoderOptions struct {
	EncodedFrameChannel chan *webrtcutil.Streamable
	CanvasChannel       <-chan *Canvas
	CloseSignal         <-chan bool
	Debugger            *debugutil.Debugger
	DimensionsChannel   <-chan *Dimensions
}

func NewEncoder(options *EncoderOptions) *Encoder {
	return &Encoder{
		encodedFrameCh:    options.EncodedFrameChannel,
		canvasCh:          options.CanvasChannel,
		closeSignal:       options.CloseSignal,
		debugger:          options.Debugger,
		dimensionsChannel: options.DimensionsChannel,
	}
}

func (e *Encoder) isClosed() bool {
	return atomic.LoadInt32(&e.closed) == 1
}

func (e *Encoder) markAsClosed() {
	atomic.StoreInt32(&e.closed, 1)
}

const ffmpegBaseCommand = "ffmpeg %v -threads 0 -re -f rawvideo -pixel_format rgb24 -video_size %dx%d -framerate %v -r %v -i pipe:0 -pix_fmt yuv420p -c:v h264_videotoolbox -f h264 pipe:1"

func (e *Encoder) Start() {

	debug := "-hide_banner -loglevel error"
	if constants.FFMPEG_BANNER {
		debug = ""
	}

	go func() {
		_, ok := <-e.closeSignal
		if !ok {
			e.markAsClosed()
			fmt.Println("Closing encoder")

			e.wg.Wait()
			e.cmd.Process.Kill()
			close(e.encodedFrameCh)
		}
	}()

	for dimensions := range e.dimensionsChannel {
		if e.cmd != nil {
			fmt.Println("DIM CHANGED")
			e.markAsClosed()

			e.cmd.Process.Kill()
			e.wg.Wait()
			fmt.Println("Closing encoder")
		}

		ffmpegCommand := fmt.Sprintf(ffmpegBaseCommand, debug, dimensions.Width, dimensions.Height, constants.FPS, constants.FPS)
		e.cmd = exec.Command("bash", "-c", ffmpegCommand)
		e.cmd.Stderr = os.Stderr
		inPipe, err := e.cmd.StdinPipe()
		logutil.LogFatal(err)
		outPipe, err := e.cmd.StdoutPipe()
		logutil.LogFatal(err)
		err = e.cmd.Start()
		logutil.LogFatal(err)

		e.wg.Add(2)

		go e.writeToFFmpeg(inPipe)
		go e.streamToWebRTCTrack(outPipe)

	}

}

func (e *Encoder) writeToFFmpeg(inPipe io.WriteCloser) {
	defer func() {
		fmt.Println("EXIT WRITE")
		e.wg.Done()
	}()

	for canvas := range e.canvasCh {
		if e.closed == 1 {
			return
		}
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
	defer func() {
		fmt.Println("EXIT STREAM")
		e.wg.Done()
	}()

	buf := make([]byte, 1024*8)
	for {
		if e.closed == 1 {
			return
		}
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
