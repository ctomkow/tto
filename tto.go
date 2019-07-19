package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"tto/database"
	"tto/filetransfer"
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

		// dump DB
		mysqlDump := database.Dump(config.Mysql.DBport, config.Mysql.DBip, config.Mysql.DBuser, config.Mysql.DBpass, config.Mysql.DBname)

		// open connection to remote server and copy dump over
		transferConnection := filetransfer.Open(config.System.Dest, config.System.Port, config.System.User, config.System.Pass)
		filetransfer.Send(transferConnection, mysqlDump, config.System.Replicate.BackupDir)

		// delete local dump
		filetransfer.Cleanup(mysqlDump)

	} else if strings.Compare(config.System.Role, "receiver") == 0 {
		// TODO: daemon that monitors folder for new dumps to restore (should write to a file the name of the last dump that was restored!)
		//db := database.Open(config.Mysql.DBport, config.Mysql.DBip, config.Mysql.DBuser, config.Mysql.DBpass, config.Mysql.DBname)
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
