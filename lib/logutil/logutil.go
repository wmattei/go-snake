package logutil

import (
	"fmt"
	"log"
	"time"
)

func LogFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func LogTimeElapsed(started time.Time, label string) {
	fmt.Println(label, time.Since(started))
}
