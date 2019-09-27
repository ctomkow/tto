package util

import (
	"testing"
	"time"
)

// test each time segment: year, month, day, hour, min, sec
func TestTimestamp_NewTimestamp(t *testing.T) {

	ts := NewTimestamp()

	now := time.Now().UTC()
	fmtTS := now.Format("20060102150405")
	year := now.Year()
	month := now.Month()
	day := now.Day()
	hour := now.Hour()
	min := now.Minute()
	sec := now.Second()

	if ts.utcTime.Year() != year {
		t.Errorf("Generated UTC year is wrong; expected, found: %d, %d", year, ts.utcTime.Year())
	}
	if ts.utcTime.Month() != month {
		t.Errorf("Generated UTC month is wrong; expected, found: %d, %d", month, ts.utcTime.Month())
	}
	if ts.utcTime.Day() != day {
		t.Errorf("Generated UTC day is wrong; expected, found: %d, %d", day, ts.utcTime.Day())
	}
	if ts.utcTime.Hour() != hour {
		t.Errorf("Generated UTC hour is wrong; expected, found: %d, %d", hour, ts.utcTime.Hour())
	}
	if ts.utcTime.Minute() != min {
		t.Errorf("Generated UTC min is wrong; expected, found: %d, %d", min, ts.utcTime.Minute())
	}
	if ts.utcTime.Second() != sec {
		t.Errorf("Generated UTC sec is wrong; expected, found: %d, %d", sec, ts.utcTime.Second())
	}
	if ts.fmtTime != fmtTS {
		t.Errorf("Formatted timestamp is wrong; expected, found: %s, %s", fmtTS, ts.fmtTime)
	}
}

func TestTimestamp_Timestamp(t *testing.T) {

	ts := NewTimestamp()
	now := time.Now().UTC()
	fmtTS := now.Format("20060102150405")

	if ts.Timestamp() != fmtTS {
		t.Errorf("Formatted timestamp is wrong; expected, found: %s, %s", fmtTS, ts.fmtTime)
	}
}
