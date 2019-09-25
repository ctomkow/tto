// Craig Tomkow
// July 24, 2019

package exec

import (
	"bytes"
	"github.com/ctomkow/tto/net"
	"os/exec"
)

func RemoteCmd(ssh *net.SSH, command string) (string, error) {

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

func LocalCmd(command []string) (string, error) {

	cmd := exec.Command(command[0], command[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return out.String(), nil
}
