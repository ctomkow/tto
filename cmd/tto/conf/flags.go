// Craig Tomkow
// August 2, 2019

package conf

import (
	"flag"
	"fmt"
)

func ParseFlags() {

	flag.Parse()
}

// set -conf flag and return pointer
func SetConfFlag() *string {

	// default conf file
	confFlagPtr := flag.String("conf", "conf.json", "name of conf file.")

	return confFlagPtr
}

func SetUserUsage(usage string, commands string, flags string) {

	flag.Usage = func() {
		fmt.Println(usage)
		fmt.Print(flags)
		fmt.Print(commands)
	}
}

func SetLogToStderr() error {

	// override glog default logging. Set to stderr so daemon managers can read the logs (docker, systemd)
	if err := flag.Set("logtostderr", "true"); err != nil {
		return err
	}

	return nil
}
