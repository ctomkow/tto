// Craig Tomkow
// August 6, 2019

package processes

import (
	"github.com/ctomkow/tto/net"
	"github.com/golang/glog"
	"github.com/ctomkow/tto/exec"
)

func GetRemoteDumps(sh *net.SSH, dbName string, workingDir string) (string, error) {

	result, err := exec.RemoteCmd(sh, "find " + workingDir + " -name *'" + dbName + "*.sql'")
	if err != nil {
		return "", err
	}

	return result, nil
}

func TransferDumpToRemote(sh *net.SSH, workingDir string, dumpName string, dumpBytes []byte) error {

	// add lock file on remote system for mysql dumpName
	_, err := exec.RemoteCmd(sh, "touch " + workingDir + "~" + dumpName + ".lock")
	if err != nil {
		return err
	}
	// TODO: if copy fails (e.g. timeout) then the remaining steps don't complete! They should!
	// copy dumpName to remote system
	if err = sh.CopyBytes(dumpBytes, dumpName, workingDir, "0600"); err != nil {
		return err
	}
	// remove lock file on remote system for mysql dumpName
	_, err = exec.RemoteCmd(sh,"rm " + workingDir + "~" + dumpName + ".lock")
	if err != nil {
		return err
	}
	// add lock file on remote system for .latest.dump
	_, err = exec.RemoteCmd(sh,"touch " + workingDir + "~.latest.dump.lock")
	if err != nil {
		return err
	}
	// update latest dumpName notes on remote system
	_, err = exec.RemoteCmd(sh,"echo " + dumpName + " > " + workingDir + ".latest.dump")
	if err != nil {
		return err
	}
	// remove lock file on remote system for .latest.dump
	_, err = exec.RemoteCmd(sh,"rm " + workingDir + "~.latest.dump.lock")
	if err != nil {
		return err
	}

	glog.Info("transferred db dump: " + dumpName)
	return nil
}

func DeleteRemoteDumps(sh *net.SSH, workingDir string, arrayOfFilenames []string) error {

	for _, filename := range arrayOfFilenames {

		_, err := exec.RemoteCmd(sh, "rm " + workingDir + filename)
		if err != nil {
			return err
		}

		glog.Info("deleted db dump: " + filename)
	}

	return nil
}
