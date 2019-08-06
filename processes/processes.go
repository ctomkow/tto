// Craig Tomkow
// August 2, 2019

package processes

import (
	"bufio"
	"errors"
	"github.com/ctomkow/tto/database"
	"github.com/ctomkow/tto/remote"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func RestoreDatabase(db *database.Database, workingDir string) (string, error) {

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
	if strings.Compare(strings.Split(latestDump, "-")[0], db.GetName()) != 0 {
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

	// TODO: error handling if database is DROP'd already... (not that it should be)
	// restoreDatabase mysqldump into database
	if err = db.Restore(workingDir + latestDump); err != nil {
		return "", err
	}

	// update .latest.restore with restored dump filename
	if err = ioutil.WriteFile(workingDir+".latest.restore", []byte(latestDump), 0600); err != nil {
		return "", err
	}

	return latestDump, nil
}

func GetRemoteDumps(sh *remote.SSH, dbName string, workingDir string) (string, error) {

	cmd := "find " + workingDir + " -name *'" + dbName + "*.sql'"

	if err := sh.NewSession(); err != nil {
		return "", err
	}
	result, err := sh.RunCommand(cmd)
	if err != nil {
		return "", err
	}

	return result, nil
}

func TransferDumpToRemote(sh *remote.SSH, workingDir string, dump string) (string, error) {

	// add lock file on remote system for mysql dump
	if err := sh.NewSession(); err != nil {
		return "", err
	}
	_, err := sh.RunCommand("touch " + workingDir + "~" + dump + ".lock")
	if err != nil {
		return "", err
	}

	// TODO: if copy fails (e.g. timeout) then the remaining steps don't complete! They should!
	// copy dump to remote system
	if err = sh.NewSession(); err != nil {
		return "", err
	}
	if err = sh.CopyFile(dump, workingDir, "0600"); err != nil {
		return "", err
	}

	// remove lock file on remote system for mysql dump
	if err = sh.NewSession(); err != nil {
		return "", err
	}
	_, err = sh.RunCommand("rm " + workingDir + "~" + dump + ".lock")
	if err != nil {
		return "", err
	}

	// add lock file on remote system for .latest.dump
	if err = sh.NewSession(); err != nil {
		return "", err
	}
	_, err = sh.RunCommand("touch " + workingDir + "~.latest.dump.lock")
	if err != nil {
		return "", err
	}

	// update latest dump notes on remote system
	if err = sh.NewSession(); err != nil {
		return "", err
	}
	_, err = sh.RunCommand("echo " + dump + " > " + workingDir + ".latest.dump")
	if err != nil {
		return "", err
	}

	// remove lock file on remote system for .latest.dump
	if err = sh.NewSession(); err != nil {
		return "", err
	}
	_, err = sh.RunCommand("rm " + workingDir + "~.latest.dump.lock")
	if err != nil {
		return "", err
	}

	// delete local dump
	if err = removeFile(workingDir + dump); err != nil {
		return "", err
	}

	return dump, nil
}

func DeleteRemoteDump(sh *remote.SSH, workingDir string, arrayOfFilenames []string) error {

	for _, elem := range arrayOfFilenames {

		cmd := "rm " + workingDir + elem

		if err := sh.NewSession(); err != nil {
			return err
		}
		_, err := sh.RunCommand(cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func fileExists(filename string) bool {

	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func removeFile(filename string) error {

	if err := os.Remove(filename); err != nil {
		return err
	}

	return nil
}
