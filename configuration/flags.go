// Craig Tomkow
// August 2, 2019

package configuration

import (
	"flag"
	"github.com/golang/glog"
)

// parse -conf flag and return as pointer
func CliFlags() *string {

	// override glog default logging to stderr so daemon managers can read the logs (docker, systemd)
	if err := flag.Set("logtostderr", "true"); err != nil {
		glog.Fatal(err)
	}
	// default conf file
	confFilePtr := flag.String("conf", "conf.json", "name of conf file.")

	flag.Parse()
	return confFilePtr
}
