package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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

func init() {

	fmt.Println("### ", "three-two-one go! v0.01")
}

func main() {

	// parse cli flags and config file
	configFile := cliFlags()
	config, _  := loadConfig(*configFile)

	if strings.Compare(config.System.Role, "sender") == 0 {
		// TODO: daemon that watches the clock based on config.System.Replicate.Interval

		// remote connection setup
		client := remote.ConnPrep(config.System.Dest, config.System.Port, config.System.User, config.System.Pass)
		client.Connect()

		// dump DB
		mysqlDump := database.Dump(config.Mysql.DBport, config.Mysql.DBip, config.Mysql.DBuser, config.Mysql.DBpass, config.Mysql.DBname)

		// TODO: add locking file (use remote.Command("touch .$mysqlDump.sql.lock")

		// open connection to remote server and copy dump over
		err := client.CopyFile(mysqlDump, config.System.Replicate.BackupDir, "0600")
		if err != nil {
			log.Fatal(err)
		}

		// TODO: remove locking file
		// TODO: update remote file with latest transfer. remote.Command("touch .latest.mysql.dump")

		defer client.Close()

		// TODO: close the local mysqldump that was opened

		// delete local dump
		Remove(mysqlDump)

	} else if strings.Compare(config.System.Role, "receiver") == 0 {
		// TODO: daemon that monitors folder for new dumps to restore (should write to a file the name of the last dump that was restored!)
		//db := database.SCPOpen(config.Mysql.DBport, config.Mysql.DBip, config.Mysql.DBuser, config.Mysql.DBpass, config.Mysql.DBname)
		//database.Restore(db, mysqlDump)
	}
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

