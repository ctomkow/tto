// Craig Tomkow
// August 2, 2019

package backup

import (
	"bufio"
	"errors"
	"github.com/ctomkow/tto/cmd/tto/db"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func Restore(dB db.DB, workingDir string) (string, error) {

	// ## .latest.dump actions

	// check if lock dumpFile exists for .latest.dump
	// retries 3 times with a 3 second sleep inbetween. Used for unfortunate timings...
	retryCount := 0
	for {
		if fileExists(workingDir + "~.latest.dump.lock") {
			retryCount++
			time.Sleep(3 * time.Second)
		} else {
			break
		}

		if retryCount == 3 {
			return "", errors.New("locked: .latest.dump is being used by another process, or lock file is stuck. Suggest manually removing ~.latest.dump.lock")
		}
	}

	// create ~.latest.dump.lock
	_, err := os.Create(workingDir + "~.latest.dump.lock")
	if err != nil {
		return "", err
	}

	// open .latest.dump and read first line
	dumpFile, err := os.Open(workingDir + ".latest.dump")
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(dumpFile)
	scanner.Scan()
	latestDump := scanner.Text()
	if err = dumpFile.Close(); err != nil {
		return "", err
	}

	// delete ~.latest.dump.lock
	if err = os.Remove(workingDir + "~.latest.dump.lock"); err != nil {
		return "", err
	}

	// ## safety check: latest dump vs configuration database name
	if strings.Compare(strings.Split(latestDump, "_-_")[0], dB.Name()) != 0 {
		// oh shit, someone is dumping one database but trying to restoreDatabase it into another one
		return "", errors.New("the dumped database does not match the one configured in the conf file")
	}

	// ## .latest.restore actions

	// open .latest.restore and read first line
	restoreFile, err := os.Open(workingDir + ".latest.restore")
	if err != nil {
		return "", err
	}
	scanner = bufio.NewScanner(restoreFile)
	scanner.Scan()
	latestRestore := scanner.Text()
	if err = restoreFile.Close(); err != nil {
		return "", err
	}

	// if dump and restoreDatabase the same, then return error
	if strings.Compare(latestDump, latestRestore) == 0 {
		return "", errors.New(".latest.dump and .latest.restore are the same")
	}

	// restore database dump into database
	fd, err := os.Open(workingDir+latestDump)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := fd.Close(); err != nil {
			glog.Error(err)
		}
	}()

	dumpReader := bufio.NewReader(fd)
	if err = dB.Restore(dumpReader); err != nil {
		return "", err
	}

	// update .latest.restore with restored dump filename
	if err = ioutil.WriteFile(workingDir+".latest.restore", []byte(latestDump), 0600); err != nil {
		return "", err
	}

	return latestDump, nil
}

func fileExists(filename string) bool {

	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
