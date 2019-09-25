// Craig Tomkow
// August 6, 2019

package backup

import (
	"github.com/ctomkow/tto/exec"
	"github.com/ctomkow/tto/net"
	"github.com/ctomkow/tto/netio"
	"github.com/golang/glog"
	"io"
)

// add lock file, copy dump over, remove lock, add lock for .latest.dump, update .latest.dump, remove lock
func ToRemote(sh *net.SSH, workingDir string, dumpName string, stdout *io.ReadCloser, ex *exec.Exec) error {

	_, err := ex.RemoteCmd(sh, "touch "+workingDir+"~"+dumpName+".lock")
	if err != nil {
		return err
	}
	if err = netio.Copy(stdout, dumpName, workingDir, "0600", ex, sh); err != nil {
		return err
	}
	_, err = ex.RemoteCmd(sh, "rm "+workingDir+"~"+dumpName+".lock")
	if err != nil {
		return err
	}
	_, err = ex.RemoteCmd(sh, "touch "+workingDir+"~.latest.dump.lock")
	if err != nil {
		return err
	}
	_, err = ex.RemoteCmd(sh, "echo "+dumpName+" > "+workingDir+".latest.dump")
	if err != nil {
		return err
	}
	_, err = ex.RemoteCmd(sh, "rm "+workingDir+"~.latest.dump.lock")
	if err != nil {
		return err
	}
	glog.Info("transferred db dump: " + dumpName)
	return nil
}

func GetBackups(sh *net.SSH, dbName string, workingDir string, ex *exec.Exec) (string, error) {

	result, err := ex.RemoteCmd(sh, "find "+workingDir+" -name *'"+dbName+"*.sql'")
	if err != nil {
		return "", err
	}

	return result, nil
}

func Delete(sh *net.SSH, workingDir string, filenames []string, ex *exec.Exec) error {

	for _, filename := range filenames {

		_, err := ex.RemoteCmd(sh, "rm "+workingDir+filename)
		if err != nil {
			return err
		}
		glog.Info("deleted db dump: " + filename)
	}

	return nil
}
