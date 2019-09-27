// 2019 Craig Tomkow.

// Package util provides supporting functions. The first useful function is for timestamps.
package util

import "time"

type timestamp struct {

	// stores the current time in UTC
	utcTime time.Time

	// stores formatted timestamp
	fmtTime string
}

// creates a new timestamp.
// Stores the current time in UTC and formats the default timestamp using golang's reference format
func MakeTimestamp() *timestamp {
	var ts timestamp
	ts.utcTime = time.Now().UTC()
	ts.fmtTime = ts.utcTime.Format("20060102150405")
	return &ts
}

// returns the formatted timestamp
func (t *timestamp) GetTimestamp() string {
	return t.fmtTime
}
