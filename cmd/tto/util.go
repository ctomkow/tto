// Craig Tomkow
// July 30, 2019
// TODO: need to re-work this. Use time.go and move this into the util package
package main

import (
	"bufio"
	"errors"
	"github.com/golang/glog"
	"sort"
	"strings"
	"time"
)

// stripPath removes the provided path from each string of the slice of strings
func stripPath(path string, filenames []string) []string {
	var results []string
	for _, filename := range filenames {
		results = append(results, strings.Replace(filename, path, "", -1))
	}
	return results
}

// sortBackups returns a sorted string slice based on the timestamp in the filename of the backup
func sortBackups(filenames []string) []string {
	var dbName string
	var timeString string
	var timestamps []time.Time

	for _, filename := range filenames {
		// grab before and after character sequence
		splitStrings, err := splitOnDelimiter("_-_", filename)
		if err != nil {
			glog.Fatal(err)
		}
		dbName = splitStrings[0]
		afterDash := splitStrings[1]

		// grab before dot but after dash
		splitStrings, err = splitOnDelimiter(".", afterDash)
		if err != nil {
			glog.Fatal(err)
		}
		timeString = splitStrings[0]

		// parse timestamp into time.Time
		timeOfDump, err := parseTimeString(timeString)
		if err != nil {
			glog.Fatal(err)
		}
		timestamps = append(timestamps, timeOfDump)
	}

	sortedTimestamps := sortTimestamps(timestamps)

	//re-compile the full filename
	var dumps []string
	for _, timestamp := range sortedTimestamps {

		dumps = append(dumps, compileBackupFilename(dbName, timestamp))
	}
	return dumps
}

// ## parse helpers ##

// compileBackupFilename returns a string of the full database backup filename
func compileBackupFilename(dbName string, timestamp time.Time) string {

	compiledString := dbName + "_-_" + timestamp.Format("20060102150405") + ".sql"
	return compiledString
}

// parseMultilineString takes a multiline string with '\n' delimiter
// returns a slice of strings
func parseMultilineString(str string) ([]string, error) {
	var strSlice []string
	scanner := bufio.NewScanner(strings.NewReader(str))
	for scanner.Scan() {
		strSlice = append(strSlice, scanner.Text())
	}
	return strSlice, nil
}

// splitOnDelimiter splits a string returning a string slice with all parts
func splitOnDelimiter(delimiter string, input string) ([]string, error) {
	if strings.Compare(delimiter, "") == 0 {
		return []string{}, errors.New("empty delimiter not allowed")
	}
	strSlice := strings.Split(input, delimiter)
	return strSlice, nil
}

// parseTimeString returns time.Time from a timestamp string
func parseTimeString(timeStr string) (time.Time, error) {
	parsedTime, err := time.Parse("20060102150405", timeStr)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime, nil
}

// sortTimestamps orders the times from oldest to newest
func sortTimestamps(timestamps []time.Time) []time.Time {
	var sortedTimestamps []time.Time

	// a hack to simplify sorting. O.O
	var unixTimeSlice []int64
	for _, timestamp := range timestamps {
		unixTimeSlice = append(unixTimeSlice, timestamp.Unix())
	}

	sort.Slice(unixTimeSlice, func(i, j int) bool { return unixTimeSlice[i] < unixTimeSlice[j] })

	for _, unixTime := range unixTimeSlice {
		sortedTimestamps = append(sortedTimestamps, time.Unix(unixTime, 0).UTC())
	}
	return sortedTimestamps
}
