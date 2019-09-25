// Craig Tomkow
// August 6, 2019

package backup

import (
	"github.com/ctomkow/tto/exec"
	"github.com/ctomkow/tto/net"
	"github.com/golang/glog"
)

// add lock file, copy dump over, remove lock, add lock for .latest.dump, update .latest.dump, remove lock
func ToRemote(sh *net.SSH, workingDir string, dumpName string, dumpBytes []byte) error {

	_, err := exec.RemoteCmd(sh, "touch "+workingDir+"~"+dumpName+".lock")
	if err != nil {
		return err
	}
	if err = sh.CopyBytes(dumpBytes, dumpName, workingDir, "0600"); err != nil {
		return err
	}
	_, err = exec.RemoteCmd(sh, "rm "+workingDir+"~"+dumpName+".lock")
	if err != nil {
		return err
	}
	_, err = exec.RemoteCmd(sh, "touch "+workingDir+"~.latest.dump.lock")
	if err != nil {
		return err
	}
	_, err = exec.RemoteCmd(sh, "echo "+dumpName+" > "+workingDir+".latest.dump")
	if err != nil {
		return err
	}
	_, err = exec.RemoteCmd(sh, "rm "+workingDir+"~.latest.dump.lock")
	if err != nil {
		return err
	}
	glog.Info("transferred db dump: " + dumpName)
	return nil
}

func GetBackups(sh *net.SSH, dbName string, workingDir string) (string, error) {

	result, err := exec.RemoteCmd(sh, "find "+workingDir+" -name *'"+dbName+"*.sql'")
	if err != nil {
		return "", err
	}

	return result, nil
}

func Delete(sh *net.SSH, workingDir string, arrayOfFilenames []string) error {

	for _, filename := range arrayOfFilenames {

		_, err := exec.RemoteCmd(sh, "rm "+workingDir+filename)
		if err != nil {
			return err
		}
		glog.Info("deleted db dump: " + filename)
	}

	return nil
}
