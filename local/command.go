// Craig Tomkow
// Sept 5, 2019

package local

import (
	"bytes"
	"os/exec"
)

func RunCommand(command []string) (string, error) {

	cmd := exec.Command(command[0], command[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

