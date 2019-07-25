// Craig Tomkow
// July 24, 2019

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	"github.com/takama/daemon"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"tto/database"
	"tto/remote"
)

type config struct {
	System struct {
		Role       string `json:"role"`
		Dest       string `json:"dest"`
		Port       string `json:"port"`
		User       string `json:"user"`
		Pass       string `json:"pass"`
		WorkingDir string `json:"working_dir"`
		Replicate struct {
			Mysql      string `json:"mysql"`
			Interval   string `json:"interval"`
		}
	}
	Mysql struct {
		DBip   string `json:"db_ip"`
		DBport string `json:"db_port"`
		DBuser string `json:"db_user"`
		DBpass string `json:"db_pass"`
		DBname string `json:"db_name"`
	}
}

type command struct {
	install bool
	remove  bool
	start   bool
	stop    bool
	status  bool
}

// Service has embedded daemon
type Service struct {
	daemon.Daemon
	restoreLock bool
}

const (
	// name of the service
	name        = "tto"
	description = "3-2-1 go!"
)

func main() {

	// parse cli flags
	configFile := cliFlags()

	// parse cli commands
	command := command{}
	command.cliCommands()

	// if service is being installed, create sample conf file; /etc/tto/conf.json if it doesn't exist
	switch {
		case command.install:

			// create config directory
			err := os.MkdirAll("/etc/tto/", os.ModePerm)
			if err != nil {
				glog.Fatal(err)
			}

			// if sample conf.json doesn't exist, create it
			if !fileExists("/etc/tto/conf.json") {
				fd, err := os.Create("/etc/tto/conf.json")
				if err != nil {
					glog.Exit(err)
				}
				defer fd.Close()

				sampleConfig := &config{}

				sampleConfig.System.Role = `[sender|receiver]`
				sampleConfig.System.Dest = `x.x.x.x`
				sampleConfig.System.Port = `22`
				sampleConfig.System.User = `username`
				sampleConfig.System.Pass = `password`
				sampleConfig.System.WorkingDir = `/opt/tto/`
				sampleConfig.System.Replicate.Mysql = `[true|false]`
				sampleConfig.System.Replicate.Interval = `[0000s|00m|00h|00d]`
				sampleConfig.Mysql.DBip = `y.y.y.y`
				sampleConfig.Mysql.DBport = `3306`
				sampleConfig.Mysql.DBuser = `username`
				sampleConfig.Mysql.DBpass = `password`
				sampleConfig.Mysql.DBname = `databaseName`

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
	}

	// TODO: if conf.json is deleted, `tto remove` fails

	// parse config
	config := config{}
	err  := config.loadConfig("/etc/tto/" + *configFile)
	if err != nil {
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

	srv, err := daemon.New(name, description)
	if err != nil {
		glog.Fatal(err)
	}

	service := &Service{srv, false}
	status, err := service.Manage(config, &command)
	if err != nil {
		glog.Fatal(err)
	}
	glog.Info(status)
	glog.Flush()
}

// Manage by daemon commands or run the daemon
func (service *Service) Manage(config config, command *command) (string, error) {

	usage := "Usage: tto install | remove | start | stop | status"

	if command.install {
		return service.Install()

	} else if command.remove {
		return service.Remove()

	} else if command.start {
		return service.Start()

	} else if command.stop {
		return service.Stop()

	} else if command.status {
		return service.Status()

	}

	// a ticker every config.System.Replicate.Interval Used by the sender
	interval, err := time.ParseDuration(config.System.Replicate.Interval)
	if err != nil {
		glog.Fatal(err)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// a file watcher monitoring .latest.dump used by the receiver
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		glog.Fatal(err)
	}
	defer watcher.Close()

	// FYI, VIM doesn't create a WRITE event, only RENAME, CHMOD, REMOVE (then breaks future watching)
	// https://github.com/fsnotify/fsnotify/issues/94#issuecomment-287456396
	err = watcher.Add(config.System.WorkingDir + ".latest.dump")
	if err != nil {
		glog.Fatal(err)
	}

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// define event variable
	var event fsnotify.Event

	// daemon work cycle for sender and receiver or interrupt by system signal
	for {
		select {

				// for sender, trigger on ticker interval value
			case timer := <-ticker.C:
				if strings.Compare(config.System.Role, "sender") == 0 {
					glog.Info("ticker interval " + timer.String())
					mysqlDump, err := config.dumpDatabase()
					if err != nil {
						glog.Error(err)
					} else {
						err = config.transferDump(mysqlDump)
						if err != nil {
							glog.Error(err)
						}
					}
				}

				// for receiver, trigger on event from watching .latest.dump
			case event = <-watcher.Events:
				if strings.Compare(config.System.Role, "receiver") == 0 {
					if triggerOnEvent(event) {
						if !service.restoreLock {
							service.restoreLock = true
							err := config.restore()
							if err != nil {
								glog.Error(err)
							}
							service.restoreLock = false
						} // else, silently skip and don't attempt to restore database as it's currently in progress
						// this also handles any double firing of watched WRITE events, that some editors create
					}
				}

				// trigger on signal
			case killSignal := <-interrupt:
				glog.Error(killSignal)

				if killSignal == os.Interrupt {
					return "", errors.New("daemon was interrupted by system signal")
				}
				return "", errors.New("daemon was killed")

		}
	}

	return usage, nil
}


// parse -conf flag and return as pointer
func cliFlags() *string {

	// override glog default logging to stderr so daemon managers can read the logs (docker, systemd)
	flag.Set("logtostderr", "true")
	// default conf file
	confFilePtr := flag.String("conf", "conf.json", "name of conf file.")

	flag.Parse()
	return confFilePtr
}

func (command *command) cliCommands() {

	if len(os.Args) > 1 {
		cmd := os.Args[1]
		switch cmd {
			case "install":
				command.install = true
			case "remove":
				command.remove = true
			case "start":
				command.start = true
			case "stop":
				command.stop = true
			case "status":
				command.status = true
		}
	}

}

func (config *config) loadConfig(filename string) error {

	// TODO: config file input validation. Depends if the app is a sender or receiver

	fd, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	jsonParser := json.NewDecoder(fd)
	err = jsonParser.Decode(&config)
	if err != nil {
		return err
	}

	return nil
}

func Remove(filename string) error {

	err := os.Remove(filename)
	if err != nil {
		return err
	}

	return nil
}

func triggerOnEvent(event fsnotify.Event) bool {

	if event.Op&fsnotify.Write == fsnotify.Write {
		return true
	}

    // TODO: add handling for a REMOVE event. e.g. need to re-create the file and re-listen for it

	return false
}

func (config config) dumpDatabase() (string, error) {

	// dump DB
	mysqlDump, err := database.Dump(config.Mysql.DBport, config.Mysql.DBip, config.Mysql.DBuser, config.Mysql.DBpass, config.Mysql.DBname)
	if err != nil {
		return "", err
	}

	return mysqlDump, nil
}

func (config config) transferDump(mysqlDump string) error {

	// connect to remote system
	client := remote.ConnPrep(config.System.Dest, config.System.Port, config.System.User, config.System.Pass)
	err := client.Connect()
	if err != nil {
		return err
	}

	// add lock file on remote system for mysql dump
	err = client.NewSession()
	if err != nil {
		return err
	}
	defer client.CloseSession()
	_, err = client.RunCommand("touch " + config.System.WorkingDir + "~" + mysqlDump + ".lock")
	if err != nil {
		return err
	}

	// copy dump to remote system
	err = client.NewSession()
	if err != nil {
		return err
	}
	defer client.CloseSession()
	err = client.CopyFile(mysqlDump, config.System.WorkingDir, "0600")
	if err != nil {
		return err
	}

	// remove lock file on remote system for mysql dump
	err = client.NewSession()
	if err != nil {
		return err
	}
	defer client.CloseSession()
	_, err = client.RunCommand("rm " + config.System.WorkingDir + "~" + mysqlDump + ".lock")
	if err != nil {
		return err
	}

	// add lock file on remote system for .latest.dump
	err = client.NewSession()
	if err != nil {
		return err
	}
	defer client.CloseSession()
	_, err = client.RunCommand("touch " + config.System.WorkingDir + "~.latest.dump.lock")
	if err != nil {
		return err
	}

	// update latest dump notes on remote system
	err = client.NewSession()
	if err != nil {
		return err
	}
	defer client.CloseSession()
	_, err = client.RunCommand("echo " + mysqlDump + " > " + config.System.WorkingDir + ".latest.dump")
	if err != nil {
		return err
	}

	// remove lock file on remote system for .latest.dump
	err = client.NewSession()
	if err != nil {
		return err
	}
	defer client.CloseSession()
	_, err = client.RunCommand("rm " + config.System.WorkingDir + "~.latest.dump.lock")
	if err != nil {
		return err
	}

	// delete local dump
	err = Remove(mysqlDump)
	if err != nil {
		return err
	}

	return nil
}

func (config config) restore() error {

	// ########### .latest.dump #############

	// check if lock dumpFile exists for .latest.dump
	if fileExists(config.System.WorkingDir + "~.latest.dump.lock") {
		return errors.New("locked: .latest.dump is being used by another process")
	}

	// create ~.latest.dump.lock
	_, err := os.Create(config.System.WorkingDir + "~.latest.dump.lock")
	if err != nil {
		return err
	}

	// open .latest.dump and read first line
	dumpFile, err := os.Open(config.System.WorkingDir + ".latest.dump")
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(dumpFile)
	scanner.Scan()
	latestDump := scanner.Text()
	err = dumpFile.Close()
	if err != nil {
		return err
	}

	// delete ~.latest.dump.lock
	err = os.Remove(config.System.WorkingDir + "~.latest.dump.lock")
	if err != nil {
		return err
	}

	// ########### .latest.restore #############

	// open .latest.restore and read first line
	restoreFile, err := os.Open(config.System.WorkingDir + ".latest.restore")
	if err != nil {
		return err
	}
	scanner = bufio.NewScanner(restoreFile)
	scanner.Scan()
	latestRestore := scanner.Text()
	err = restoreFile.Close()
	if err != nil {
		return err
	}

	// if dump and restore not the same, then attempt to restore the latestDump
	if strings.Compare(latestDump, latestRestore) != 0 {

		// TODO: error handling if database is DROP'd already... (not that it should be)

		// open connection to database
		db, err := database.Open(config.Mysql.DBport, config.Mysql.DBip, config.Mysql.DBuser, config.Mysql.DBpass, config.Mysql.DBname)
		if err != nil {
			return err
		}

		// restore mysqldump into database
		err = database.Restore(db, config.System.WorkingDir+ latestDump)
		if err != nil {
			return err
		}

		// update .latest.restore with restored dump filename
		err = ioutil.WriteFile(config.System.WorkingDir+ ".latest.restore", []byte(latestDump), 0600)
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New(".latest.dump and .latest.restore are the same")
}

func fileExists(filename string) bool {

	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

