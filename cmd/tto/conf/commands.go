// Craig Tomkow
// August 2, 2019

package conf

import (
	"errors"
	"flag"
)

type Command struct {
	Install bool
	Remove  bool
	Fg      bool
}

func (cmd *Command) MakeCmd() error {

	if len(flag.Args()) > 1 {
		return errors.New("only one command allowed, or flags should be before the command. See --help for more info")
	}

	switch flag.Arg(0) {
	case "install":
		cmd.Install = true
	case "remove":
		cmd.Remove = true
	case "fg":
		cmd.Fg = true
	default:
		return errors.New("invalid command: " + flag.Arg(0))
	}

	return nil
}
