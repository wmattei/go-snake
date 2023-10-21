package encodingutil_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/pion/webrtc/v3/pkg/media/h264reader"
	"github.com/wmattei/go-snake/constants"
	"github.com/wmattei/go-snake/shared/logutil"
)

func generateFrame() []byte {
	rawRGBData := make([]byte, 3*constants.FRAME_WIDTH*constants.FRAME_HEIGHT)
	idx := 0
	for y := 0; y < 768; y++ {
		for x := 0; x < 1024; x++ {
			rawRGBData[idx] = 255
			rawRGBData[idx+1] = 255
			rawRGBData[idx+2] = 255
			idx += 3
		}
	}

	return rawRGBData
}

func BenchmarkEncoding(b *testing.B) {
	rawFrame := generateFrame()

	for i := 0; i < b.N; i++ {

		ffmpegCommand := "ffmpeg -hide_banner -loglevel error -f rawvideo -pixel_format rgb24 -video_size 1024x768 -framerate 1 -i pipe:0 -c:v libx264 -preset ultrafast -tune zerolatency -f h264 pipe:1"

		cmd := exec.Command("bash", "-c", ffmpegCommand)
		cmd.Stderr = os.Stderr

		inPipe, err := cmd.StdinPipe()
		logutil.LogFatal(err)
		outPipe, err := cmd.StdoutPipe()
		logutil.LogFatal(err)

		if err := cmd.Start(); err != nil {
			b.Fail()
		}

		_, err = inPipe.Write(rawFrame)
		if err != nil {
			b.Fail()
		}

		inPipe.Close()
		encodedData, err := readH264NALUnits(outPipe)
		if err != nil {
			b.Fail()
		}

		err = cmd.Wait()
		if err != nil {
			b.Fail()
		}
		if len(encodedData) == 0 {
			b.Fail()
		}
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
