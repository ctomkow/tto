package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/takama/daemon"
	"log"
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
		Role      string `json:"role"`
		Dest      string `json:"dest"`
		Port      string `json:"port"`
		User      string `json:"user"`
		Pass      string `json:"pass"`
		Replicate struct {
			Mysql     string `json:"mysql"`
			Interval  string `json:"interval"`
			BackupDir string `json:"backup_dir"`
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

func init() {

	// TODO: handling to ensure .latest.dump and .latest.restore files exist in working directory!


	// TODO: deal with all log.Fatal()'s throughout the codebase
}

func main() {

	// parse cli flags and config file
	configFile := cliFlags()
	config, _  := loadConfig(*configFile)

	srv, err := daemon.New(name, description)
	if err != nil {
		log.Fatal(err)
	}

	service := &Service{srv, false}
	status, err := service.Manage(config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(status)
}

// Manage by daemon commands or run the daemon
func (service *Service) Manage(config config) (string, error) {

	usage := "Usage: tto install | remove | start | stop | status"

	// if received any kind of command, do it
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}

	// a ticker every config.System.Replicate.Interval Used by the sender
	interval, err := time.ParseDuration(config.System.Replicate.Interval)
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// a file watcher monitoring .latest.dump Used by the receiver
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	// TODO: FYI, VIM doesn't create a WRITE event, only RENAME, CHMOD, REMOVE (then breaks future watching)
	//   https://github.com/fsnotify/fsnotify/issues/94#issuecomment-287456396
	err = watcher.Add(config.System.Replicate.BackupDir + ".latest.dump")
	if err != nil {
		log.Fatal(err)
	}

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// loop work cycle for sender and receiver or interrupt by system signal
	for {
		select {

			// for sender, trigger on ticker interval value
			case timer := <- ticker.C:
				if strings.Compare(config.System.Role, "sender") == 0 {
					fmt.Println(timer)
					config.transferDump(config.dumpDatabase())
				}

			// for receiver, trigger on event from watching .latest.dump
			case event, _ := <- watcher.Events:
				if strings.Compare(config.System.Role, "receiver") == 0 {
					if triggerOnEvent(event) == true {
						if service.restoreLock == false {
							service.restoreLock = true
							err := config.restore()
							if err != nil {
								log.Fatal(err)
							}
							service.restoreLock = false
						}
					}
				}

			// for everything, trigger on signal
			case killSignal := <-interrupt:
				fmt.Println("Got signal:", killSignal)

				if killSignal == os.Interrupt {
					return "Daemon was interrupted by system signal", nil
				}
				return "Daemon was killed", nil
		}
	}

	// never happen, but need to complete code
	return usage, nil
}

// parse -conf flag and return as pointer
func cliFlags() *string {

	confFilePtr := flag.String("conf", "conf.json", "name of conf file.")
	flag.Parse()
	return confFilePtr
}

func loadConfig(filename string) (config, error) {

	var configStruct config
	fd, err := os.Open(filename)
	defer fd.Close() // close fd when function returns

	if err != nil {
		log.Fatal(err)
	}

	jsonParser := json.NewDecoder(fd)
	err = jsonParser.Decode(&configStruct)
	return configStruct, err
}

func Remove(filename string) {

	err := os.Remove(filename)
	if err != nil {
		log.Fatal(err)
	}
}

func triggerOnEvent(event fsnotify.Event) bool {

	if event.Op&fsnotify.Write == fsnotify.Write {
		return true
	}
    // TODO: add handling for a REMOVE event. e.g. need to re-create the file and re-listen for it

	return false
}

func (config config) dumpDatabase() string {

	// dump DB
	mysqlDump := database.Dump(config.Mysql.DBport, config.Mysql.DBip, config.Mysql.DBuser, config.Mysql.DBpass, config.Mysql.DBname)

	return mysqlDump
}

func (config config) transferDump(mysqlDump string) {

	// connect to remote system
	client := remote.ConnPrep(config.System.Dest, config.System.Port, config.System.User, config.System.Pass)
	client.Connect()

	// TODO: add handling for if .lock already exists (fileExists)

	// add lock file on remote system for mysql dump
	client.NewSession()
	client.RunCommand("touch " + config.System.Replicate.BackupDir + "~" + mysqlDump + ".lock")
	client.CloseSession()

	// copy dump to remote system
	client.NewSession()
	err := client.CopyFile(mysqlDump, config.System.Replicate.BackupDir, "0600")
	if err != nil {
		log.Fatal(err)
	}
	client.CloseSession()

	// remove lock file on remote system for mysql dump
	client.NewSession()
	client.RunCommand("rm " + config.System.Replicate.BackupDir + "~" + mysqlDump + ".lock")
	client.CloseSession()

	// TODO: add handling for if .lock already exists (fileExists)

	// add lock file on remote system for .latest.dump
	client.NewSession()
	client.RunCommand("touch " + config.System.Replicate.BackupDir + "~.latest.dump.lock")
	client.CloseSession()

	// update latest dump notes on remote system
	client.NewSession()
	client.RunCommand("echo " + mysqlDump + " > " + config.System.Replicate.BackupDir + ".latest.dump")
	client.CloseSession()

	// remove lock file on remote system for .latest.dump
	client.NewSession()
	client.RunCommand("rm " + config.System.Replicate.BackupDir + "~.latest.dump.lock")
	client.CloseSession()

	// delete local dump
	Remove(mysqlDump)
}

func (config config) restore() error {

	// ########### .dump.lock #############

	// check if lock dumpFile exists for .latest.dump
	if fileExists(config.System.Replicate.BackupDir + "~.latest.dump.lock") {
		return errors.New("locked: .latest.dump is being used by another process")
	}

	// TODO: create ~.latest.dump.lock

	// open .latest.dump and read first line
	dumpFile, err := os.Open(config.System.Replicate.BackupDir + ".latest.dump")
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(dumpFile)
	scanner.Scan()
	latestDump := scanner.Text()
	dumpFile.Close()

	// TODO: delete ~.latest.dump.lock

	// ########### .restore.lock #############

	// check if lock dumpFile exists for .latest.restore
	if fileExists(config.System.Replicate.BackupDir + "~.latest.restore.lock") {
		return errors.New("locked: .latest.restore is being used by another process")
	}

	// TODO: create ~.latest.restore.lock

	// open .latest.restore and read first line
	restoreFile, err := os.Open(config.System.Replicate.BackupDir + ".latest.restore")
	if err != nil {
		log.Fatal(err)
	}

	scanner = bufio.NewScanner(restoreFile)
	scanner.Scan()
	latestRestore := scanner.Text()

	// if dump and restore not the same, then attempt to restore the latestDump
	if strings.Compare(latestDump, latestRestore) != 0 {
		// TODO: error handling if database is dropped already... (not that it should be)

		// open connection to database
		db := database.Open(config.Mysql.DBport, config.Mysql.DBip, config.Mysql.DBuser, config.Mysql.DBpass, config.Mysql.DBname)

		// TODO: add error handling to ensure that if a database.Restore fails, I am not accidentally still writing to .latest.restore
		// restore mysqldump into database
		database.Restore(db, config.System.Replicate.BackupDir + latestDump)

		// TODO: ugghh, writing to file is not working. look into this
		// update .latest.restore with restored dump filename
		_, _ = restoreFile.WriteString(latestDump + "\n")
		restoreFile.Close()

		// TODO: delete ~.latest.restore.lock

		return nil
	}

	return errors.New("database restore of " + config.System.Replicate.BackupDir + latestDump + "failed")
}

func fileExists(filename string) bool {

	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

