// Craig Tomkow
// July 24, 2019

package remote

import (
	"bytes"
)

// ##### public methods #####

func (sc *SSH) RunCommand(command string) (string, error) {

	var stdoutBuffer bytes.Buffer
	sc.session.Stdout = &stdoutBuffer
	if err := sc.session.Run(command); err != nil {
		return "", err
	}

	return stdoutBuffer.String(), nil
}
