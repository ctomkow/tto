// Craig Tomkow
// August 6, 2019

package processes

import (
	"github.com/ctomkow/tto/remote"
	"github.com/golang/glog"
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

func TransferDumpToRemote(sh *remote.SSH, workingDir string, dumpName string, dumpBytes []byte) error {

	// add lock file on remote system for mysql dumpName
	if err := sh.NewSession(); err != nil {
		return err
	}
	_, err := sh.RunCommand("touch " + workingDir + "~" + dumpName + ".lock")
	if err != nil {
		return err
	}

	// TODO: if copy fails (e.g. timeout) then the remaining steps don't complete! They should!
	// copy dumpName to remote system
	if err = sh.NewSession(); err != nil {
		return err
	}
	if err = sh.CopyBytes(dumpBytes, dumpName, workingDir, "0600"); err != nil {
		return err
	}

	// remove lock file on remote system for mysql dumpName
	if err = sh.NewSession(); err != nil {
		return err
	}
	_, err = sh.RunCommand("rm " + workingDir + "~" + dumpName + ".lock")
	if err != nil {
		return err
	}

	// add lock file on remote system for .latest.dump
	if err = sh.NewSession(); err != nil {
		return err
	}
	_, err = sh.RunCommand("touch " + workingDir + "~.latest.dump.lock")
	if err != nil {
		return err
	}

	// update latest dumpName notes on remote system
	if err = sh.NewSession(); err != nil {
		return err
	}
	_, err = sh.RunCommand("echo " + dumpName + " > " + workingDir + ".latest.dump")
	if err != nil {
		return err
	}

	// remove lock file on remote system for .latest.dump
	if err = sh.NewSession(); err != nil {
		return err
	}
	_, err = sh.RunCommand("rm " + workingDir + "~.latest.dump.lock")
	if err != nil {
		return err
	}

	glog.Info("transferred db dump: " + dumpName)
	return nil
}

func DeleteRemoteDumps(sh *remote.SSH, workingDir string, arrayOfFilenames []string) error {

	for _, filename := range arrayOfFilenames {

		cmd := "rm " + workingDir + filename

		if err := sh.NewSession(); err != nil {
			return err
		}
		_, err := sh.RunCommand(cmd)
		if err != nil {
			return err
		}
		glog.Info("deleted db dump: " + filename)
	}

	return nil
}
