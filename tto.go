// Craig Tomkow
// July 24, 2019

package main

import (
	"encoding/json"
	"errors"
	"github.com/ctomkow/tto/configuration"
	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	"github.com/takama/daemon"
	"os"
	"os/user"
	"strconv"
)

type Service struct {
	daemon.Daemon
}

const (
	// name of the service
	name        = "tto"
	description = "3-2-1 go!"
)

func main() {

	// parse cli flags
	configFile := configuration.CliFlags()

	// parse cli commands
	var cmd = new(configuration.Command)
	cmd.MakeCmd()

	// if service is being installed, create sample conf file; /etc/tto/conf.json if it doesn't exist
	switch {
	case cmd.Install:

		// create config directory if it doesn't exist
		if err := os.MkdirAll("/etc/tto/", os.ModePerm); err != nil {
			glog.Fatal(err)
		}

		// if sample conf.json doesn't exist, create it
		if !fileExists("/etc/tto/conf.json") {
			fd, err := os.Create("/etc/tto/conf.json")
			if err != nil {
				glog.Exit(err)
			}
			defer func() {
				if err := fd.Close(); err != nil {
					glog.Exit(err)
				}
			}()

			// populate with sample configuration
			var sampleConf = new(configuration.Config)
			sampleConf.MakeConfig()


			var jsonData []byte
			jsonData, err = json.MarshalIndent(sampleConf, "", "    ")
			if err != nil {
				glog.Error(err)
			}

			_, err = fd.WriteString(string(jsonData))
			if err != nil {
				glog.Error(err)
			}
			glog.Info("created file: /etc/tto/conf.json")
		}

		// create working directory if it doesn't exist
		if err := os.MkdirAll("/opt/tto/", os.ModePerm); err != nil {
			glog.Fatal(err)
		}
	}

	// TODO: if conf.json is deleted, `tto remove` fails

	// parse config
	var conf = new(configuration.Config)
	if err := conf.LoadConfig("/etc/tto/" + *configFile); err != nil {
		glog.Exit(err)
	}

	// ensure working directory files exists
	if !fileExists(conf.System.WorkingDir + ".latest.dump") {
		_, err := os.Create(conf.System.WorkingDir + ".latest.dump")
		if err != nil {
			glog.Exit(err)
		}
		glog.Info("created file: " + conf.System.WorkingDir + ".latest.dump")
	}

	// ensure working directory files exists
	if !fileExists(conf.System.WorkingDir + ".latest.restore") {
		_, err := os.Create(conf.System.WorkingDir + ".latest.restore")
		if err != nil {
			glog.Exit(err)
		}
		glog.Info("created file: " + conf.System.WorkingDir + ".latest.restore")
	}

	// chown all files to appropriate usr

	// get app uid/gid based on system.conf from conf.json
	usr, err := user.Lookup(conf.System.User)
	if err != nil {
		glog.Exit(err)
	}
	uid, _ := strconv.Atoi(usr.Uid)
	gid, _ := strconv.Atoi(usr.Gid)

	if err = os.Chown("/opt/tto/", uid, gid); err != nil {
		glog.Exit(err)
	}

	if err = os.Chown("/opt/tto/.latest.dump", uid, gid); err != nil {
		glog.Exit(err)
	}

	if err = os.Chown("/opt/tto/.latest.restore", uid, gid); err != nil {
		glog.Exit(err)
	}

	// TODO: run service as a user. The daemon package should set this in the systemd file!

	// what is my role
	daemonRole := conf.System.Type

	// daemon setup and service start
	srv, err := daemon.New(name, description)
	if err != nil {
		glog.Fatal(err)
	}

	service := &Service{srv}
	status, err := service.Manage(conf, cmd, daemonRole)
	if err != nil {
		glog.Fatal(err)
	}
	glog.Info(status)
	glog.Flush()
}

// daemon manager

func (srv *Service) Manage(conf *configuration.Config, cmd *configuration.Command, role string) (string, error) {

	usage := "Usage: tto install | remove | start | stop | status"

	if cmd.Install {
		return srv.Install()

	} else if cmd.Remove {
		return srv.Remove()

	} else if cmd.Start {
		return srv.Start()

	} else if cmd.Stop {
		return srv.Stop()

	} else if cmd.Status {
		return srv.Status()

	}

	switch role {
	case "sender":
		if err := Sender(conf); err != nil {
			return "", err
		}

	case "receiver":
		if err := Receiver(conf); err != nil {
			return "", err
		}

	default:
		return "", errors.New("could not start daemon! unknown type: " + role)
	}

	return usage, nil
}

func cronTriggered(c chan bool) {

	c <- true
}

func isWriteEvent(event fsnotify.Event) bool {

	if event.Op&fsnotify.Write == fsnotify.Write {
		return true
	}

	return false
}

func fileExists(filename string) bool {

	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
