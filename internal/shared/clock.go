package shared

import "time"

// Clock abstracts the time provider to enable time-freezing
// and deterministic time generation in tests (PKS-029).
type Clock interface {
	// Now returns the current local time.
	Now() time.Time
	// After waits for the duration to elapse and then sends the current time on the returned channel.
	After(d time.Duration) <-chan time.Time
}

// RealClock is the production implementation of the Clock interface.
type RealClock struct{}

func NewRealClock() *RealClock {
	return &RealClock{}
}

func (c *RealClock) Now() time.Time {
	return time.Now()
}

func (c *RealClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}
