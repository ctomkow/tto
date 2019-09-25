// Craig Tomkow
// July 24, 2019

package exec

import (
	"bytes"
	"errors"
	"github.com/ctomkow/tto/db"
	"github.com/ctomkow/tto/net"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
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

	// YYYYMMDDhhmmss
	currentTime := time.Now().UTC().Format("20060102150405") //TODO: remove static time format (or move it), buffer also relies on this format

	ipArg := "-h" + db.Ip.String()
	portArg := "-P" + strconv.FormatUint(uint64(db.Port), 10)
	userArg := "-u" + db.Username
	passArg := "-p" + db.Password
	sqlFile := db.Name + "-" + currentTime + ".sql"

	if strings.Compare(db.Impl, "mysql") == 0 {
		c.Cmd = exec.Command("mysqldump", "--single-transaction", "--skip-lock-tables", "--routines", "--triggers", ipArg, portArg, userArg, passArg, db.Name)
	} else {
		return nil, "", errors.New("unsupported database type")
	}

	stdout, err := c.Cmd.StdoutPipe()
	if err != nil {
		return nil, "", err
	}

	if err = c.Cmd.Start(); err != nil {
		return nil, "", err
	}

	return &stdout, sqlFile, nil
}
