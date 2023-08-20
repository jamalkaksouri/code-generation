package utils

import (
	"fmt"
	"time"
)

func CalculateRemainingTime(elapsedTime time.Duration, currentIteration, remainingIterations int) time.Duration {
	averageIterationTime := elapsedTime / time.Duration(currentIteration)
	return averageIterationTime * time.Duration(remainingIterations)
}

func FormatDuration(d time.Duration) string {
	const (
		secondsInMinute = 60
		minutesInHour   = 60
		hoursInDay      = 24
	)

	days := int(d.Hours()) / hoursInDay
	hours := int(d.Hours()) % hoursInDay
	minutes := int(d.Minutes()) % minutesInHour
	seconds := int(d.Seconds()) % secondsInMinute

	if days > 0 {
		return fmt.Sprintf("%d day %02d hour", days, hours)
	} else if hours > 0 {
		return fmt.Sprintf("%02d hour %02d minute", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%02d minute %02d second", minutes, seconds)
	} else {
		return fmt.Sprintf("%02d second", seconds)
	}
}
