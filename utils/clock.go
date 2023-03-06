package utils

import "time"

type Clock struct {
	MockTime time.Time
}

func (c Clock) Now() time.Time {
	if c.MockTime.IsZero() {
		return time.Now()
	} else {
		return c.MockTime
	}
}
