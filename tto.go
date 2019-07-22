package main

import (
	"encoding/json"
	"flag"
	"fmt"
	//"github.com/fsnotify/fsnotify"
	"log"
	"github.com/takama/daemon"
	//"net"
	"os"
	"os/signal"
	"syscall"
	//"time"
	//"tto/database"
	//"tto/remote"
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

const (
	// name of the service
	name        = "tto"
	description = "3-2-1 go!"
)

// Service has embedded daemon
type Service struct {
	daemon.Daemon
}

// Manage by daemon commands or run the daemon
func (service *Service) Manage() (string, error) {

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

	// TODO: time setup.
	//certainSomething := true // will cause time loop to repeat
	//timeDelay := 900 * time.Millisecond // == 900000 * time.Microsecond
	//var endTime <-chan time.Time // signal for when timer us up

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// loop work cycle with accept connections or interrupt
	// by system signal
	for {
		select {
		// TODO: when time is triggered, receive notice from time channel. Trigger function that has a goroutine in it
		//case conn := <-listen:
		//	go handleClient(conn)
		// TODO: add another case, where when a file is changed, receive notice from file-change channel, trigger function with goroutine to restore database
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

func init() {

	// TODO: handling to ensure latest.dump and latest.restore files exist in working directory!
}

func main() {

	// parse cli flags and config file
	configFile := cliFlags()
	config, _  := loadConfig(*configFile)

	srv, err := daemon.New(name, description)
	if err != nil {
		log.Fatal(err)
	}

	service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(status)

	fmt.Println(config)

	/*

	if strings.Compare(config.System.Role, "sender") == 0 {
		// TODO: daemon that watches the clock based on config.System.Replicate.Interval

		// connect to remote system
		client := remote.ConnPrep(config.System.Dest, config.System.Port, config.System.User, config.System.Pass)
		client.Connect()

		// dump DB
		mysqlDump := database.Dump(config.Mysql.DBport, config.Mysql.DBip, config.Mysql.DBuser, config.Mysql.DBpass, config.Mysql.DBname)

		// add lock file on remote system
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

		// remove lock file
		client.NewSession()
		client.RunCommand("rm " + config.System.Replicate.BackupDir + "~" + mysqlDump + ".lock")
		client.CloseSession()

		// update latest dump notes on remote system
		client.NewSession()
		client.RunCommand("echo " + mysqlDump + " > " + config.System.Replicate.BackupDir + "latest.dump")
		client.CloseSession()

		// delete local dump
		Remove(mysqlDump)

	} else if strings.Compare(config.System.Role, "receiver") == 0 {

		// create file watcher
		//watcher, err := fsnotify.NewWatcher()
		//if err != nil {
		//	log.Fatal(err)
		//}
		//
		//err = watcher.Add(config.System.Replicate.BackupDir + "latest.dump")

		//database.Restore(db, mysqlDump)

	}

	*/

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

