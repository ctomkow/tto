// Craig Tomkow
// July 30, 2019

package ringbuffer

import (
	"bufio"
	"errors"
	"github.com/golang/glog"
	"sort"
	"strings"
	"time"
)

// ring buffer prep work: count existing backups, store into ring buffer

func Parse(dumps string) []time.Time {

	remoteFilesArray, err := parseMultilineString(dumps)

	var slicedElem []string
	var parsedTimeSlice []time.Time

	for _, elem := range remoteFilesArray {

		// grab after dash
		slicedElem, err = sliceOnDelimiter("-", elem)
		if err != nil {
			glog.Fatal(err)
		}

		// grab before dot
		slicedElem, err = sliceOnDelimiter(".", slicedElem[1])
		if err != nil {
			glog.Fatal(err)
		}

		// parse timestamp into time.Time
		timeOfDump, err := parseTimeString(slicedElem[0])
		if err != nil {
			glog.Fatal(err)
		}

		parsedTimeSlice = append(parsedTimeSlice, timeOfDump)
	}

	parsedTimeSlice = sortTimeSlice(parsedTimeSlice)

	return parsedTimeSlice
}

// ## parse helpers ##

func parseMultilineString(mString string) ([]string, error) {

	var strArry []string

	scanner := bufio.NewScanner(strings.NewReader(mString))
	for scanner.Scan() {
		strArry = append(strArry, scanner.Text())
	}

	return strArry, nil
}

func sliceOnDelimiter(delimiter string, concatStr string) ([]string, error) {

	if strings.Compare(delimiter, "") == 0 {
		return []string{}, errors.New("empty delimiter not allowed")

	}
	strSlice := strings.Split(concatStr, delimiter)

	return strSlice, nil
}

func parseTimeString(timeStr string) (time.Time, error) {

	parsedTime, err := time.Parse("20060102150405", timeStr)
	if err != nil {
		return time.Time{}, err
	}

	return parsedTime, nil
}

func sortTimeSlice(timeSlice []time.Time) []time.Time {

	// a hack to simplify sorting. O.O
	var unixTimeSlice []int64
	for _, elem := range timeSlice {
		unixTimeSlice = append(unixTimeSlice, elem.Unix())
	}

	sort.Slice(unixTimeSlice, func(i, j int) bool { return unixTimeSlice[i] < unixTimeSlice[j] })

	var newTimeSlice []time.Time
	for _, elem := range unixTimeSlice {
		newTimeSlice = append(newTimeSlice, time.Unix(elem, 0).UTC())
	}

	return newTimeSlice
}

func compileFilename(dbName string, fileTime time.Time) string {

	var compiledString string
	compiledString = dbName + "-" + fileTime.Format("20060102150405") + ".sql"
	return compiledString
}
