package clock

import "time"

// Clock provides time for event timestamps
// This abstraction allows using virtual time in tests for deterministic simulation
type Clock interface {
	Now() time.Time
}

// RealClock uses the actual system time
type RealClock struct{}

// Now returns the current system time
func (r *RealClock) Now() time.Time {
	return time.Now()
}

// VirtualClock allows manual control of time for testing
type VirtualClock struct {
	current time.Time
}

// NewVirtualClock creates a virtual clock starting at the given time
func NewVirtualClock(start time.Time) *VirtualClock {
	return &VirtualClock{current: start}
}

// Now returns the current virtual time
func (v *VirtualClock) Now() time.Time {
	return v.current
}

// Set sets the virtual time to a specific value
func (v *VirtualClock) Set(t time.Time) {
	v.current = t
}

// Advance moves the virtual time forward by the given duration
func (v *VirtualClock) Advance(d time.Duration) {
	v.current = v.current.Add(d)
}
