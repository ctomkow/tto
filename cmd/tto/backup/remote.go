// Craig Tomkow
// August 6, 2019

package backup

import (
	"github.com/ctomkow/tto/cmd/tto/exec"
	"github.com/ctomkow/tto/cmd/tto/inet"
	"github.com/ctomkow/tto/cmd/tto/netio"
	"github.com/golang/glog"
	"io"
)

// add lock file, copy dump over, remove lock, add lock for .latest.dump, update .latest.dump, remove lock
func ToRemote(sh *inet.SSH, workingDir string, dumpName string, stdout *io.ReadCloser, ex *exec.Exec) error {

	_, err := ex.RemoteCmd(sh, "touch "+workingDir+"~"+dumpName+".lock")
	if err != nil {
		return err
	}
	if err = netio.StreamMySqlDump(stdout, dumpName, workingDir, "0600", ex, sh); err != nil {
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

// Retrieve returns a multiline string of database dumps that is delimited based on the remote host's operating system
func Retrieve(sh *inet.SSH, exe *exec.Exec, dbName string, workingDir string) (string, error) {
	result, err := exe.RemoteCmd(sh, "find "+workingDir+" -name *'"+dbName+"*.sql'")
	if err != nil {
		return "", err
	}
	return result, nil
}

// Delete removes files from a remote host
func Delete(sh *inet.SSH, exe *exec.Exec, workingDir string, filenames []string) error {
	for _, filename := range filenames {
		_, err := exe.RemoteCmd(sh, "rm "+workingDir+filename)
		if err != nil {
			return err
		}
		glog.Info("deleted db dump: " + filename)
	}
	return nil
}
