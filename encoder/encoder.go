package encoder

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/pion/webrtc/v3/pkg/media/h264reader"
	"github.com/wmattei/go-snake/constants"
	"github.com/wmattei/go-snake/snake_errors"
)

func StartEncoder(pixelCh chan []byte, encodedFrameCh chan []byte) {
	const ffmpegCommand = "ffmpeg -hide_banner -loglevel error -f rawvideo -pixel_format rgb24 -video_size %dx%d -framerate %d -i pipe:0 -c:v libx264 -preset ultrafast -tune zerolatency -f h264 pipe:1"
	for {
		rawRGBDataFrame := <-pixelCh
		if rawRGBDataFrame == nil {
			break
		}
		cmd := exec.Command("bash", "-c", fmt.Sprintf(ffmpegCommand, constants.FRAME_WIDTH, constants.FRAME_HEIGHT, constants.FPS))
		cmd.Stderr = os.Stderr

		inPipe, err := cmd.StdinPipe()
		snake_errors.HandleError(err)

		outPipe, err := cmd.StdoutPipe()
		snake_errors.HandleError(err)

		err = cmd.Start()
		snake_errors.HandleError(err)

		_, err = inPipe.Write(rawRGBDataFrame)
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
		cmd.Wait()
		encodedFrameCh <- data
	}
}
