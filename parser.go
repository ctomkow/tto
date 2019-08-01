// Craig Tomkow
// July 30, 2019

package main

import (
	"bufio"
	"errors"
	"github.com/golang/glog"
	"sort"
	"strings"
	"time"
)

// ring buffer prep work: count existing backups, store into ring buffer

func ParseDbDumpFilename(filename string) []time.Time {

	arrayOfStrings, err := parseMultilineString(filename)
	if err != nil {
		glog.Fatal(err)
	}

	var arrayOfTimestamps []time.Time

	for _, elem := range arrayOfStrings {

		// grab after dash
		slicedElements, err := sliceOnDelimiter("-", elem)
		if err != nil {
			glog.Fatal(err)
		}
		afterDash := slicedElements[1]

		// grab before dot
		slicedElements, err = sliceOnDelimiter(".", afterDash)
		if err != nil {
			glog.Fatal(err)
		}
		betweenDashDot := slicedElements[0]

		// parse timestamp into time.Time
		timeOfDump, err := parseTimeString(betweenDashDot)
		if err != nil {
			glog.Fatal(err)
		}

		arrayOfTimestamps = append(arrayOfTimestamps, timeOfDump)
	}

	arrayOfTimestamps = sortTimestamps(arrayOfTimestamps)

	return arrayOfTimestamps
}

func CompileDbDumpFilename(dbName string, timestamp time.Time) string {

	compiledString := dbName + "-" + timestamp.Format("20060102150405") + ".sql"
	return compiledString
}

// ## parse helpers ##

func parseMultilineString(mString string) ([]string, error) {

	var strArray []string

	scanner := bufio.NewScanner(strings.NewReader(mString))
	for scanner.Scan() {
		strArray = append(strArray, scanner.Text())
	}

	return strArray, nil
}

func sliceOnDelimiter(delimiter string, inputString string) ([]string, error) {

	if strings.Compare(delimiter, "") == 0 {
		return []string{}, errors.New("empty delimiter not allowed")

	}
	strSlice := strings.Split(inputString, delimiter)

	return strSlice, nil
}

func parseTimeString(timeStr string) (time.Time, error) {

	parsedTime, err := time.Parse("20060102150405", timeStr)
	if err != nil {
		return time.Time{}, err
	}

	return parsedTime, nil
}

func sortTimestamps(timestamps []time.Time) []time.Time {

	// a hack to simplify sorting. O.O
	var unixTimeSlice []int64
	for _, elem := range timestamps {
		unixTimeSlice = append(unixTimeSlice, elem.Unix())
	}

	sort.Slice(unixTimeSlice, func(i, j int) bool { return unixTimeSlice[i] < unixTimeSlice[j] })

	var sortedTimestamps []time.Time
	for _, elem := range unixTimeSlice {
		sortedTimestamps = append(sortedTimestamps, time.Unix(elem, 0).UTC())
	}

	return sortedTimestamps
}
