package utils

import (
	"time"
)

var startTime time.Time

func GetEndDelay() time.Duration {
	return time.Minute*10 - time.Since(startTime)
}

func SetStartTime(t time.Time) {
	startTime = t
}

func WaitGameTime(d time.Duration) <-chan time.Time {
	return time.After(d - time.Since(startTime))
}
