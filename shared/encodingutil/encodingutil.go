package encodingutil

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"github.com/pion/webrtc/v3/pkg/media/h264reader"
	"github.com/wmattei/go-snake/shared/logutil"
)

const ffmpegBaseCommand = "ffmpeg -hide_banner -loglevel error -f rawvideo -pixel_format rgb24 -video_size %dx%d -framerate 1 -i pipe:0 -c:v libx264 -preset ultrafast -tune zerolatency -f h264 pipe:1"

var ffmpegCommand string

func encodeFrame(rawFrame []byte, windowWidth, windowHeight int) ([]byte, error) {

	cmd := exec.Command("bash", "-c", ffmpegCommand)
	cmd.Stderr = os.Stderr

	inPipe, err := cmd.StdinPipe()
	logutil.LogFatal(err)
	outPipe, err := cmd.StdoutPipe()
	logutil.LogFatal(err)

	if err := cmd.Start(); err != nil {
		logutil.LogFatal(err)
		return nil, err
	}

	_, err = inPipe.Write(rawFrame)
	if err != nil {
		return nil, err
	}

	inPipe.Close()

	encodedData, err := readH264NALUnits(outPipe)
	if err != nil {
		return nil, err
	}

	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	return encodedData, nil
}

func StartEncoder(pixelCh chan []byte, encodedFrameCh chan []byte, windowWidth, windowHeight int) {
	ffmpegCommand = fmt.Sprintf(ffmpegBaseCommand, windowWidth, windowHeight)

	numWorkers := runtime.NumCPU()

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for {
				rawRGBDataFrame, more := <-pixelCh
				if !more {
					// Channel closed, exit the loop
					break
				}

				encodedData, err := encodeFrame(rawRGBDataFrame, windowWidth, windowHeight)
				logutil.LogFatal(err)
				encodedFrameCh <- encodedData
			}
		}()
	}
}

func readH264NALUnits(outPipe io.Reader) ([]byte, error) {
	h264, err := h264reader.NewReader(outPipe)
	if err != nil {
		return nil, fmt.Errorf("failed to create H.264 reader: %v", err)
	}

	var data []byte
	var spsAndPpsCache []byte

	for {
		nal, h264Err := h264.NextNAL()
		if h264Err == io.EOF {
			// Finished sending frames
			break
		} else if h264Err != nil {
			return nil, fmt.Errorf("error reading H.264 NAL: %v", h264Err)
		}

		nal.Data = append([]byte{0x00, 0x00, 0x00, 0x01}, nal.Data...)
		if nal.UnitType == h264reader.NalUnitTypeSPS || nal.UnitType == h264reader.NalUnitTypePPS {
			spsAndPpsCache = append(spsAndPpsCache, nal.Data...)
			continue
		} else if nal.UnitType == h264reader.NalUnitTypeCodedSliceIdr {
			nal.Data = append(spsAndPpsCache, nal.Data...)
			spsAndPpsCache = []byte{}
		}

		// Append NAL unit data to the result
		data = append(data, nal.Data...)
	}

	return data, nil
}
