// Craig Tomkow
// July 24, 2019

package main

import (
	"errors"
	"github.com/ctomkow/tto/configuration"
	"github.com/golang/glog"
	"github.com/takama/daemon"
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

	configFile := configuration.CliFlags()

	var cmd = new(configuration.Command)
	cmd.MakeCmd()

	if cmd.Install {
		install()
	}

	// daemon setup and service start
	srv, err := daemon.New(name, description)
	if err != nil {
		glog.Fatal(err)
	}

	service := &Service{srv}
	status, err := service.Manage(cmd, configFile)
	if err != nil {
		glog.Fatal(err)
	}
	glog.Info(status)
	glog.Flush()
}

// daemon manager

func (srv *Service) Manage(cmd *configuration.Command, configFile *string) (string, error) {

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

	var conf = new(configuration.Config)
	if err := conf.LoadConfig("/etc/tto/" + *configFile); err != nil {
		glog.Exit(err)
	}

	setupWorkingDir(conf)
	setupPermissions(conf)

	switch conf.System.Type {
	case "sender":
		if err := Sender(conf); err != nil {
			return "", err
		}

	case "receiver":
		if err := Receiver(conf); err != nil {
			return "", err
		}

	default:
		return "", errors.New("could not start daemon! unknown type: " + conf.System.Type)
	}

	return usage, nil
}
