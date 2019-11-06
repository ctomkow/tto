// Craig Tomkow
// July 24, 2019

package exec

import (
	"bytes"
	"github.com/ctomkow/tto/inet"
	"os/exec"
)

type Exec struct {

	// currently executing command
	Cmd *exec.Cmd
}

func (c *Exec) RemoteCmd(ssh *inet.SSH, command string) (string, error) {

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

// set pointer to the running command. Mainly used for streaming database dumps
func (c *Exec) LocalCmdOnly(command []string) {
	c.Cmd = exec.Command(command[0], command[1:]...)
}
