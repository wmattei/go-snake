package internal

import (
	"time"
)

func Debounce[K any](input <-chan K, duration time.Duration) <-chan K {
	output := make(chan K)
	go func() {
		var timer *time.Timer
		var latestMsg K
		for {
			select {
			case msg := <-input:
				latestMsg = msg
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(duration, func() {
					output <- latestMsg
				})
			}
		}
	}()
	return output
}
