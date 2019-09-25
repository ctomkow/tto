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

func StripPath(path string, filenames []string) []string {

	var results []string
	for _, elem := range filenames {

		results = append(results, strings.Replace(elem, path, "", -1))
	}

	return results
}

func SortBackups(filenames []string) []string {

	var dbName string
	var timeString string

	var arrayOfTimestamps []time.Time
	for _, elem := range filenames {

		// grab before and after dash
		slicedElements, err := sliceOnDelimiter("-", elem)
		if err != nil {
			glog.Fatal(err)
		}
		dbName    = slicedElements[0]
		afterDash := slicedElements[1]

		// grab before dot but after dash
		slicedElements, err = sliceOnDelimiter(".", afterDash)
		if err != nil {
			glog.Fatal(err)
		}
		timeString = slicedElements[0]

		// parse timestamp into time.Time
		timeOfDump, err := parseTimeString(timeString)
		if err != nil {
			glog.Fatal(err)
		}

		arrayOfTimestamps = append(arrayOfTimestamps, timeOfDump)
	}

	arrayOfTimestamps = sortTimestamps(arrayOfTimestamps)

	//re-compile the full filename
	var dumps []string
	for _, elem := range arrayOfTimestamps {

		dumps = append(dumps, compileBackupFilename(dbName, elem))
	}

	return dumps
}

// ## parse helpers ##

func compileBackupFilename(dbName string, timestamp time.Time) string {

	compiledString := dbName + "-" + timestamp.Format("20060102150405") + ".sql"
	return compiledString
}

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
