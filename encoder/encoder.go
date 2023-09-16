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

const ffmpegCommand = "ffmpeg -hide_banner -loglevel error -f rawvideo -pixel_format rgb24 -video_size %dx%d -framerate %d -i pipe:0 -c:v libx264 -preset ultrafast -tune zerolatency -f h264 pipe:1"

func StartEncoder(pixelCh chan []byte, encodedFrameCh chan []byte) {
	for {
		rawRGBDataFrame, ok := <-pixelCh
		if !ok {
			// Channel closed, exit the loop
			break
		}

		cmd := exec.Command("bash", "-c", fmt.Sprintf(ffmpegCommand, constants.FRAME_WIDTH, constants.FRAME_HEIGHT, constants.FPS))
		cmd.Stderr = os.Stderr

		// Create a pipe for input and output
		inPipe, err := cmd.StdinPipe()
		snake_errors.HandleError(err)
		outPipe, err := cmd.StdoutPipe()
		snake_errors.HandleError(err)

		// Start the command
		if err := cmd.Start(); err != nil {
			snake_errors.HandleError(err)
			continue
		}

		// Write raw RGB data to the input pipe
		_, err = inPipe.Write(rawRGBDataFrame)
		snake_errors.HandleError(err)

		// Close the input pipe to indicate no more input
		inPipe.Close()

		// Read H.264 NAL units from the output pipe and send to the channel
		encodedData, err := readH264NALUnits(outPipe)
		snake_errors.HandleError(err)

		// Wait for the command to finish
		err = cmd.Wait()
		snake_errors.HandleError(err)

		// Send the encoded data to the channel
		encodedFrameCh <- encodedData
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
