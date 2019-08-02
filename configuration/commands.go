// Craig Tomkow
// August 2, 2019

package configuration

import "os"

type Command struct {
	Install bool
	Remove  bool
	Start   bool
	Stop    bool
	Status  bool
}

func (cmd *Command) MakeCmd() {

	if len(os.Args) > 1 {
		cmds := os.Args[1]
		switch cmds {
		case "install":
			cmd.Install = true
		case "remove":
			cmd.Remove = true
		case "start":
			cmd.Start = true
		case "stop":
			cmd.Stop = true
		case "status":
			cmd.Status = true
		}
	}

}
