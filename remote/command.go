package remote

import (
	"log"
)

func (sc *SSH) RunCommand(command string) {

	err := sc.session.Run(command)
	if err != nil {
		log.Fatal(err)
	}
}