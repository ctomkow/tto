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
	usage 		= "Usage: [flags] (install | remove | start | stop | status | fg)"
	flags		=
		`
	--help
		prints this message
	--conf string
		custom named configuration file. default is conf.json
	`
	commands	=
	`
	install
		creates a daemon manager script depending on the service manager (SysV, Systemd, runit)
	remove
		deletes the daemon manager script that was installed
	start
		runs the daemon in the background
	stop
		gracefully stops the daemon
	status
		displays whether the daemon is running or not
	fg
		runs the program in the foreground. Needed for process managers to manage the app (docker, supervisord)
	`
)

func main() {

	if err := configuration.SetLogToStderr(); err != nil {
		glog.Fatal(err)
	}
	configFile := configuration.SetConfFlag()
	configuration.SetUserUsage(usage, commands,  flags)
	configuration.ParseFlags()

	var cmd = new(configuration.Command)
	if err := cmd.MakeCmd(); err != nil {
		glog.Fatal(err)
	}

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

	} else if cmd.Fg {
		// pass through
		glog.Info("running in foreground")
	} else {
		glog.Fatal(usage)
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
