package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/pion/webrtc/v3/pkg/media/h264reader"
	"github.com/wmattei/go-snake/constants"
	"github.com/wmattei/go-snake/shared/errutil"
)

func generateFrame() []byte {
	rawRGBData := make([]byte, 3*constants.FRAME_WIDTH*constants.FRAME_HEIGHT)
	idx := 0
	for y := 0; y < constants.FRAME_HEIGHT; y++ {
		for x := 0; x < constants.FRAME_WIDTH; x++ {
			rawRGBData[idx] = 255
			rawRGBData[idx+1] = 255
			rawRGBData[idx+2] = 255
			idx += 3
		}
	}

	return rawRGBData
}
func main() {
	frame := generateFrame()

	const outputMP4File = "output.mp4"
	const ffmpegCommand = "ffmpeg -hide_banner -loglevel error -y -f rawvideo -pixel_format rgb24 -video_size %dx%d -i pipe:0 -f h264 pipe:1"

	cmd := exec.Command("bash", "-c", fmt.Sprintf(ffmpegCommand, constants.FRAME_WIDTH, constants.FRAME_HEIGHT))

	cmd.Stderr = os.Stderr

	inPipe, err := cmd.StdinPipe()
	errutil.HandleError(err)

	outPipe, err := cmd.StdoutPipe()
	errutil.HandleError(err)

	err = cmd.Start()
	errutil.HandleError(err)

	_, err = inPipe.Write(frame)
	errutil.HandleError(err)

	inPipe.Close()

	h264, err := h264reader.NewReader(outPipe)
	errutil.HandleError(err)

	spsAndPpsCache := []byte{}
	for {
		nal, h264Err := h264.NextNAL()
		if h264Err == io.EOF {
			fmt.Println("All video frames parsed and sent")
			break
		}
		if h264Err != nil {
			panic(h264Err)
		}
		nal.Data = append([]byte{0x00, 0x00, 0x00, 0x01}, nal.Data...)

		if nal.UnitType == h264reader.NalUnitTypeSPS || nal.UnitType == h264reader.NalUnitTypePPS {
			spsAndPpsCache = append(spsAndPpsCache, nal.Data...)
			continue
		} else if nal.UnitType == h264reader.NalUnitTypeCodedSliceIdr {
			nal.Data = append(spsAndPpsCache, nal.Data...)
			spsAndPpsCache = []byte{}
		}

		fmt.Println(nal.Data)
	}

	outPipe.Close()

	cmd.Wait()
}
