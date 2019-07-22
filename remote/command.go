package remote

import (
	"bytes"
	"log"
)

func (sc *SSH) RunCommand(command string) string {

	var stdoutBuffer bytes.Buffer
	sc.session.Stdout = &stdoutBuffer
	err := sc.session.Run(command)
	if err != nil {
		log.Fatal(err)
	}

	return stdoutBuffer.String()
}
