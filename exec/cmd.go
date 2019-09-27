// Craig Tomkow
// July 24, 2019

package exec

import (
	"bytes"
	"github.com/ctomkow/tto/db"
	"github.com/ctomkow/tto/net"
	"github.com/ctomkow/tto/util"
	"io"
	"os/exec"
	"strconv"
)

type Exec struct {

	// currently executing command
	Cmd *exec.Cmd
}

func (c *Exec) RemoteCmd(ssh *net.SSH, command string) (string, error) {

	// ensure a new session is created before acting!
	if err := ssh.NewSession(); err != nil {
		return "", err
	}

	sh := ssh.GetSession()

	var stdoutBuffer bytes.Buffer
	sh.Stdout = &stdoutBuffer
	if err := sh.Run(command); err != nil {
		return "", err
	}

	return stdoutBuffer.String(), nil
}

func (c *Exec) LocalCmd(command []string) (string, error) {

	cmd := exec.Command(command[0], command[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func (c *Exec) MySqlDump(db *db.Database, workingDir string) (*io.ReadCloser, string, error) {

	timestamp := util.MakeTimestamp().GetTimestamp()
	ipArg     := "-h" + db.Ip.String()
	portArg   := "-P" + strconv.FormatUint(uint64(db.Port), 10)
	userArg   := "-u" + db.Username
	passArg   := "-p" + db.Password
	filename  := db.Name + "-" + timestamp + ".sql"

	c.Cmd = exec.Command("mysqldump", "--single-transaction", "--skip-lock-tables", "--routines", "--triggers", ipArg, portArg, userArg, passArg, db.Name)

	stdout, err := c.Cmd.StdoutPipe()
	if err != nil {
		return nil, "", err
	}

	if err = c.Cmd.Start(); err != nil {
		return nil, "", err
	}

	return &stdout, filename, nil
}
