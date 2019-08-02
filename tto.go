// Craig Tomkow
// July 24, 2019

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	"github.com/takama/daemon"
	"net"
	"os"
	"os/user"
	"strconv"
)

type config struct {
	System struct {
		User       string `json:"user"`
		Pass       string `json:"pass"`
		WorkingDir string `json:"working_dir"`
		Type       string `json:"type"`
		Role       struct {
			Sender struct {
				Dest       net.IPAddr `json:"dest"`
				Port       uint16     `json:"port"`
				Database   string     `json:"database"`
				DBip       net.IPAddr `json:"db_ip"`
				DBport     uint16     `json:"db_port"`
				DBuser     string     `json:"db_user"`
				DBpass     string     `json:"db_pass"`
				DBname     string     `json:"db_name"`
				Cron       string     `json:"cron"`
				MaxBackups int        `json:"max_backups"`
			}
			Receiver struct {
				Database string     `json:"database"`
				DBip     net.IPAddr `json:"db_ip"`
				DBport   uint16     `json:"db_port"`
				DBuser   string     `json:"db_user"`
				DBpass   string     `json:"db_pass"`
				DBname   string     `json:"db_name"`
			}
		}
	}
}

type command struct {
	install bool
	remove  bool
	start   bool
	stop    bool
	status  bool
}

type Service struct {
	daemon.Daemon
}

const (
	// name of the service
	name        = "tto"
	description = "3-2-1 go!"
)

func (cmd *command) cliCmds() {

	if len(os.Args) > 1 {
		cmds := os.Args[1]
		switch cmds {
		case "install":
			cmd.install = true
		case "remove":
			cmd.remove = true
		case "start":
			cmd.start = true
		case "stop":
			cmd.stop = true
		case "status":
			cmd.status = true
		}
	}
}

func (conf *config) loadConfig(filename string) error {

	// TODO: config file input validation. Depends if the app is a sender or receiver

	fd, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := fd.Close(); err != nil {
			glog.Exit(err)
		}
	}()

	jsonParser := json.NewDecoder(fd)
	if err = jsonParser.Decode(&conf); err != nil {
		return err
	}

	return nil
}

func main() {

	// parse cli flags
	configFile := cliFlags()

	// parse cli commands
	cmd := command{}
	cmd.cliCmds()

	// if service is being installed, create sample conf file; /etc/tto/conf.json if it doesn't exist
	switch {
	case cmd.install:

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

			sampleConfig := &config{}

			sampleConfig.System.User = `username`
			sampleConfig.System.Pass = `password`
			sampleConfig.System.WorkingDir = `/opt/tto/`
			sampleConfig.System.Type = `sender|receiver`
			sampleConfig.System.Role.Sender.Dest = net.IPAddr{net.IPv4(6, 6, 6, 6), ""}
			sampleConfig.System.Role.Sender.Port = uint16(22)
			sampleConfig.System.Role.Sender.Database = `mysql`
			sampleConfig.System.Role.Sender.DBip = net.IPAddr{net.IPv4(7, 7, 7, 7), ""}
			sampleConfig.System.Role.Sender.DBport = uint16(3306)
			sampleConfig.System.Role.Sender.DBuser = `username`
			sampleConfig.System.Role.Sender.DBpass = `password`
			sampleConfig.System.Role.Sender.DBname = `databaseName`
			sampleConfig.System.Role.Sender.Cron = `a cron statement`
			sampleConfig.System.Role.Sender.MaxBackups = int(5)
			sampleConfig.System.Role.Receiver.Database = `mysql`
			sampleConfig.System.Role.Receiver.DBip = net.IPAddr{net.IPv4(8, 8, 8, 8), ""}
			sampleConfig.System.Role.Receiver.DBport = uint16(3306)
			sampleConfig.System.Role.Receiver.DBuser = `username`
			sampleConfig.System.Role.Receiver.DBpass = `password`
			sampleConfig.System.Role.Receiver.DBname = `databaseName`

			var jsonData []byte
			jsonData, err = json.MarshalIndent(sampleConfig, "", "    ")
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
	config := config{}
	if err := config.loadConfig("/etc/tto/" + *configFile); err != nil {
		glog.Exit(err)
	}

	// ensure working directory files exists
	if !fileExists(config.System.WorkingDir + ".latest.dump") {
		_, err := os.Create(config.System.WorkingDir + ".latest.dump")
		if err != nil {
			glog.Exit(err)
		}
		glog.Info("created file: " + config.System.WorkingDir + ".latest.dump")
	}

	// ensure working directory files exists
	if !fileExists(config.System.WorkingDir + ".latest.restore") {
		_, err := os.Create(config.System.WorkingDir + ".latest.restore")
		if err != nil {
			glog.Exit(err)
		}
		glog.Info("created file: " + config.System.WorkingDir + ".latest.restore")
	}

	// chown all files to appropriate usr

	// get app uid/gid based on system.conf from conf.json
	usr, err := user.Lookup(config.System.User)
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
	daemonRole := config.System.Type

	// daemon setup and service start
	srv, err := daemon.New(name, description)
	if err != nil {
		glog.Fatal(err)
	}

	service := &Service{srv}
	status, err := service.Manage(config, &cmd, daemonRole)
	if err != nil {
		glog.Fatal(err)
	}
	glog.Info(status)
	glog.Flush()
}

// daemon manager

func (srv *Service) Manage(conf config, cmd *command, role string) (string, error) {

	usage := "Usage: tto install | remove | start | stop | status"

	if cmd.install {
		return srv.Install()

	} else if cmd.remove {
		return srv.Remove()

	} else if cmd.start {
		return srv.Start()

	} else if cmd.stop {
		return srv.Stop()

	} else if cmd.status {
		return srv.Status()

	}

	switch role {
	case "sender":
		if err := conf.Sender(); err != nil {
			return "", err
		}

	case "receiver":
		if err := conf.Receiver(); err != nil {
			return "", err
		}

	default:
		return "", errors.New("could not start daemon! unknown type: " + role)
	}

	return usage, nil
}

// ## app init helpers ##

// parse -conf flag and return as pointer
func cliFlags() *string {

	// override glog default logging to stderr so daemon managers can read the logs (docker, systemd)
	if err := flag.Set("logtostderr", "true"); err != nil {
		glog.Fatal(err)
	}
	// default conf file
	confFilePtr := flag.String("conf", "conf.json", "name of conf file.")

	flag.Parse()
	return confFilePtr
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
