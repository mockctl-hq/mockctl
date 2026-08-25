package shared

import "time"

// Clock abstracts the time provider to enable time-freezing
// and deterministic time generation in tests (PKS-029).
type Clock interface {
	// Now returns the current local time.
	Now() time.Time
}

// RealClock is the production implementation of the Clock interface.
type RealClock struct{}

func NewRealClock() *RealClock {
	return &RealClock{}
}

func (c *RealClock) Now() time.Time {
	return time.Now()
}
