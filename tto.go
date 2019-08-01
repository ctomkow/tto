// Craig Tomkow
// July 24, 2019

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"github.com/ctomkow/tto/database"
	"github.com/ctomkow/tto/remote"
	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	"github.com/robfig/cron"
	"github.com/takama/daemon"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ##### structs #####

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
				MaxBackups int       `json:"max_backups"`
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
	restoreLock bool
}

// ##### constants #####

const (
	// name of the service
	name        = "tto"
	description = "3-2-1 go!"
)

// ##### methods #####

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
	defer func() {
		if err := fd.Close(); err != nil {
			glog.Exit(err)
		}
	}()

	jsonParser := json.NewDecoder(fd)
	if err = jsonParser.Decode(&config); err != nil {
		return err
	}

	return nil
}

// ##### main #####

func main() {

	// parse cli flags
	configFile := cliFlags()

	// parse cli commands
	command := command{}
	command.cliCommands()

	// if service is being installed, create sample conf file; /etc/tto/conf.json if it doesn't exist
	switch {
	case command.install:

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
			sampleConfig.System.Role.Sender.Dest = net.IPAddr{net.IPv4(6,6,6,6), ""}
			sampleConfig.System.Role.Sender.Port = uint16(22)
			sampleConfig.System.Role.Sender.Database = `mysql`
			sampleConfig.System.Role.Sender.DBip = net.IPAddr{net.IPv4(7,7,7,7), ""}
			sampleConfig.System.Role.Sender.DBport = uint16(3306)
			sampleConfig.System.Role.Sender.DBuser = `username`
			sampleConfig.System.Role.Sender.DBpass = `password`
			sampleConfig.System.Role.Sender.DBname = `databaseName`
			sampleConfig.System.Role.Sender.Cron = `a cron statement`
			sampleConfig.System.Role.Sender.MaxBackups = int(5)
			sampleConfig.System.Role.Receiver.Database = `mysql`
			sampleConfig.System.Role.Receiver.DBip = net.IPAddr{net.IPv4(8,8,8,8), ""}
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

	service := &Service{srv, false}
	status, err := service.Manage(config, &command, daemonRole)
	if err != nil {
		glog.Fatal(err)
	}
	glog.Info(status)
	glog.Flush()
}

// ##### daemon manager #####

func (service *Service) Manage(config config, command *command, role string) (string, error) {

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

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	switch role {
	case "sender":

		// setup database connection for sender
		var db = new(database.Database)
		db.Make(config.System.Role.Sender.Database,
			config.System.Role.Sender.DBip,
			config.System.Role.Sender.DBport,
			config.System.Role.Sender.DBuser,
			config.System.Role.Sender.DBpass,
			config.System.Role.Sender.DBname)

		// get remote files
		remoteFiles, err := config.getRemoteDumps(config.System.Role.Sender.DBname)
		if err != nil {
			glog.Fatal(err)
		}

		// init ring buffer with existing files
		var buff = new(CircularQueue)
		sortedTimeSlice := ParseDbDumpFilename(remoteFiles)
		numOfBackups := config.System.Role.Sender.MaxBackups
		if err != nil {
			glog.Fatal(err)
		}
		buffOverflowTimestamps := buff.Make(numOfBackups, config.System.Role.Sender.DBname, sortedTimeSlice)
		glog.Info(errors.New("ring buffer filled up to max_backups with existing database dumps"))

		// delete any remote files that don't fit into ring buffer
		if err := config.deleteRemoteDump(config.System.Role.Sender.DBname, buffOverflowTimestamps); err != nil {
			glog.Error(err)
		}
		glog.Info(errors.New("database dumps on remote machine that didn't fit in ring buffer have been deleted"))

		// cron setup
		cronChannel := make(chan bool)
		cj := cron.New()
		err = cj.AddFunc(config.System.Role.Sender.Cron, func() { cronTriggered(cronChannel) })
		if err != nil {
			glog.Fatal(err)
		}
		cj.Start()

		for {
			select {

			// cron trigger
			case <-cronChannel:
				mysqlDump, err := db.Dump(config.System.WorkingDir)
				if err != nil {
					glog.Error(err)
				}
				if err == nil {
					copiedDump, err := config.transferDumpToRemote(mysqlDump)
					if err != nil {
						glog.Error(err)
					}
					glog.Info(errors.New("dumped and copied over database: " + copiedDump))

					// add to ring buffer and delete any overwritten file
					buffOverflowTimestamp := buff.Enqueue(config.System.Role.Sender.DBname, ParseDbDumpFilename(mysqlDump)[0])
					if !buffOverflowTimestamp.IsZero() {
						if err := config.deleteRemoteDump(config.System.Role.Sender.DBname, []time.Time{buffOverflowTimestamp}); err != nil {
							glog.Error(err)
						}
						glog.Info(errors.New("deleted old database dump: " + CompileDbDumpFilename(config.System.Role.Sender.DBname, buffOverflowTimestamp)))
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

	case "receiver":

		// setup database connection for receiver
		var db = new(database.Database)
		db.Make(config.System.Role.Sender.Database,
			config.System.Role.Sender.DBip,
			config.System.Role.Sender.DBport,
			config.System.Role.Sender.DBuser,
			config.System.Role.Sender.DBpass,
			config.System.Role.Sender.DBname)

		// a file watcher monitoring .latest.dump used by the receiver
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			glog.Fatal(err)
		}
		defer func() {
			if err := watcher.Close(); err != nil {
				glog.Exit(err)
			}
		}()

		// FYI, VIM doesn't create a WRITE event, only RENAME, CHMOD, REMOVE (then breaks future watching)
		// https://github.com/fsnotify/fsnotify/issues/94#issuecomment-287456396
		if err = watcher.Add(config.System.WorkingDir + ".latest.dump"); err != nil {
			glog.Fatal(err)
		}
		var event fsnotify.Event

		// create channel for communicating with the database restoreDatabase routine
		restoreChan := make(chan string)

		for {
			select {

			// trigger on write event
			case event = <-watcher.Events:
				if isWriteEvent(event) {
					if !service.restoreLock {

						service.restoreLock = true

						// run restoreDatabase as a goroutine. goroutine holds a restoreDatabase lock until it's done
						go func() {
							restoredDump, err := restoreDatabase(db, config.System.WorkingDir)
							if err != nil {
								glog.Error(err)
								restoreChan <- ""
							}
							restoreChan <- restoredDump
						}()

					} // else, silently skip and don't attempt to restoreDatabase database as it's currently in progress
				}

			// trigger on dump restoreDatabase being finished
			case restoredDump := <-restoreChan:
				service.restoreLock = false
				if restoredDump == "" {
					glog.Error(errors.New("failed to restoreDatabase database"))
				} else {
					glog.Info(errors.New("restored database: " + restoredDump))
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

	default:
		return "", errors.New("could not start daemon! unknown type: " + role)
	}

	return usage, nil
}

// ##### helper functions #####

// ## database helpers ##

func restoreDatabase(db *database.Database, workingDir string) (string, error) {

	// ## .latest.dump actions

	// check if lock dumpFile exists for .latest.dump
	// retries 3 times with a 3 second sleep inbetween. Used for unfortunate timings...
	retryCount := 0
	for {
		if fileExists(workingDir + "~.latest.dump.lock") {
			retryCount++
			time.Sleep(3 * time.Second)
		} else {
			break
		}

		if retryCount == 3 {
			return "", errors.New("locked: .latest.dump is being used by another process, or lock file is stuck. Suggest manually removing ~.latest.dump.lock")
		}
	}

	// create ~.latest.dump.lock
	_, err := os.Create(workingDir + "~.latest.dump.lock")
	if err != nil {
		return "", err
	}

	// open .latest.dump and read first line
	dumpFile, err := os.Open(workingDir + ".latest.dump")
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(dumpFile)
	scanner.Scan()
	latestDump := scanner.Text()
	if err = dumpFile.Close(); err != nil {
		return "", err
	}

	// delete ~.latest.dump.lock
	if err = os.Remove(workingDir + "~.latest.dump.lock"); err != nil {
		return "", err
	}

	// ## safety check: latest dump vs configuration database name
	if strings.Compare(strings.Split(latestDump, "-")[0], db.GetName()) != 0 {
		// oh shit, someone is dumping one database but trying to restoreDatabase it into another one
		return "", errors.New("the dumped database does not match the one configured in the conf file")
	}

	// ## .latest.restore actions

	// open .latest.restore and read first line
	restoreFile, err := os.Open(workingDir + ".latest.restore")
	if err != nil {
		return "", err
	}
	scanner = bufio.NewScanner(restoreFile)
	scanner.Scan()
	latestRestore := scanner.Text()
	if err = restoreFile.Close(); err != nil {
		return "", err
	}

	// if dump and restoreDatabase not the same, then attempt to restoreDatabase the latestDump
	if strings.Compare(latestDump, latestRestore) != 0 {

		// TODO: error handling if database is DROP'd already... (not that it should be)
		// restoreDatabase mysqldump into database
		if err = db.Restore(workingDir+latestDump); err != nil {
			return "", err
		}

		// update .latest.restore with restored dump filename
		if err = ioutil.WriteFile(workingDir+".latest.restore", []byte(latestDump), 0600); err != nil {
			return "", err
		}

		return latestDump, nil
	}

	return "", errors.New(".latest.dump and .latest.restore are the same")
}

// ## remote system helpers ##

func (config config) getRemoteDumps(dbName string) (string, error) {

	cmd := "find " + config.System.WorkingDir + " -name *" + dbName + "*"

	// connect to remote system
	client := remote.ConnPrep(
		config.System.Role.Sender.Dest.String(),
		strconv.FormatUint(uint64(config.System.Role.Sender.Port), 10),
		config.System.User,
		config.System.Pass)
	if err := client.Connect(); err != nil {
		return "", err
	}
	if err := client.NewSession(); err != nil {
		return "", err
	}
	result, err := client.RunCommand(cmd)
	if err != nil {
		return "", err
	}
	if err = client.CloseConnection(); err != nil {
		return "", err
	}

	return result, nil
}

func (config config) transferDumpToRemote(mysqlDump string) (string, error) {

	// connect to remote system
	client := remote.ConnPrep(
		config.System.Role.Sender.Dest.String(),
		strconv.FormatUint(uint64(config.System.Role.Sender.Port), 10),
		config.System.User,
		config.System.Pass)
	if err := client.Connect(); err != nil {
		return "", err
	}

	// add lock file on remote system for mysql dump
	if err := client.NewSession(); err != nil {
		return "", err
	}
	_, err := client.RunCommand("touch " + config.System.WorkingDir + "~" + mysqlDump + ".lock")
	if err != nil {
		return "", err
	}

	// copy dump to remote system
	if err = client.NewSession(); err != nil {
		return "", err
	}
	if err = client.CopyFile(mysqlDump, config.System.WorkingDir, "0600"); err != nil {
		return "", err
	}

	// remove lock file on remote system for mysql dump
	if err = client.NewSession(); err != nil {
		return "", err
	}
	_, err = client.RunCommand("rm " + config.System.WorkingDir + "~" + mysqlDump + ".lock")
	if err != nil {
		return "", err
	}

	// add lock file on remote system for .latest.dump
	if err = client.NewSession(); err != nil {
		return "", err
	}
	_, err = client.RunCommand("touch " + config.System.WorkingDir + "~.latest.dump.lock")
	if err != nil {
		return "", err
	}

	// update latest dump notes on remote system
	if err = client.NewSession(); err != nil {
		return "", err
	}
	_, err = client.RunCommand("echo " + mysqlDump + " > " + config.System.WorkingDir + ".latest.dump")
	if err != nil {
		return "", err
	}

	// remove lock file on remote system for .latest.dump
	if err = client.NewSession(); err != nil {
		return "", err
	}
	_, err = client.RunCommand("rm " + config.System.WorkingDir + "~.latest.dump.lock")
	if err != nil {
		return "", err
	}

	// delete local dump
	if err = removeFile(config.System.WorkingDir + mysqlDump); err != nil {
		return "", err
	}

	return mysqlDump, nil
}

func (config config) deleteRemoteDump(dbName string, arrayOfTimestamps []time.Time) error {

	// connect to remote system
	client := remote.ConnPrep(
		config.System.Role.Sender.Dest.String(),
		strconv.FormatUint(uint64(config.System.Role.Sender.Port), 10),
		config.System.User,
		config.System.Pass)
	if err := client.Connect(); err != nil {
		return err
	}

	for _, elem := range arrayOfTimestamps {

		cmd := "rm " + config.System.WorkingDir + CompileDbDumpFilename(dbName, elem)

		if err := client.NewSession(); err != nil {
			return err
		}
		_, err := client.RunCommand(cmd)
		if err != nil {
			return err
		}
	}

	if err := client.CloseConnection(); err != nil {
		return err
	}

	return nil
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

// ## cron helpers ##

func cronTriggered(c chan bool) {

	c <- true
}

// ## event helpers ##

func isWriteEvent(event fsnotify.Event) bool {

	if event.Op&fsnotify.Write == fsnotify.Write {
		return true
	}

	return false
}

// ## file helpers ##

func fileExists(filename string) bool {

	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func removeFile(filename string) error {

	if err := os.Remove(filename); err != nil {
		return err
	}

	return nil
}
