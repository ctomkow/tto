// Craig Tomkow
// August 6, 2019

package processes

import (
	"github.com/ctomkow/tto/remote"
	"os"
)

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

func removeFile(filename string) error {

	if err := os.Remove(filename); err != nil {
		return err
	}

	return nil
}
