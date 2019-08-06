// Craig Tomkow
// July 24, 2019

package remote

import (
	"bytes"
)

func (sh *SSH) RunCommand(command string) (string, error) {

	var stdoutBuffer bytes.Buffer
	sh.session.Stdout = &stdoutBuffer
	if err := sh.session.Run(command); err != nil {
		return "", err
	}

	return stdoutBuffer.String(), nil
}
