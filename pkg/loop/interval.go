package loop

import (
	"math"
	"time"
)

// CalculateInterval Calculate the interval time based on the number of times the loop has been executed.
func (l *Loop) CalculateInterval(loopedTimes uint64) time.Duration {
	// if loopedTimes == 0 Means the first execution of the task, the time interval should be 0
	if loopedTimes == 0 {
		return time.Duration(0)
	}
	// interval = initialInterval * declineRatio^(loopedTimes-1)
	// change loopedTimes+1 to loopedTimes-1 to make interval is accurate
	interval := time.Duration(float64(l.interval) * math.Pow(l.declineRatio, float64(loopedTimes-1)))
	// interval not exceed declineLimit
	if interval > l.declineLimit {
		interval = l.declineLimit
	}
	// interval not less than initial interval
	if interval < 0 {
		interval = time.Duration(math.Min(float64(l.declineLimit), float64(l.interval)))
	}
	return interval
}
