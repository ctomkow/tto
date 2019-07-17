package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"tto/database"
)

type config struct {
	System struct {
		Role      string `json:"role"`
		Replicate struct {
			Mysql    string `json:"mysql"`
			Interval string `json:"interval"`
		}
	}
	Mysql struct {
		DBport string `json:"db_port"`
		DBip   string `json:"db_ip"`
		DBuser string `json:"db_user"`
		DBpass string `json:"db_pass"`
		DBname string `json:"db_name"`
	}
}

func init() {

	fmt.Println("### ", "three-two-one sync and backup!")
	fmt.Println("### ", "version 0.01")
	fmt.Println("###")
}

func main() {

	// parse cli flag input
	configFile := cliFlags()

	fmt.Println(time.Now().Format(time.RFC850))

	// parse config file input
	config, _ := loadConfig(*configFile)

	// database connection
	db := database.ConnectToDatabase(config.Mysql.DBport, config.Mysql.DBip, config.Mysql.DBuser, config.Mysql.DBpass, config.Mysql.DBname)

	// dump DB, return dump file name
	mysqlDump := database.DumpDatabase(config.Mysql.DBport, config.Mysql.DBip, config.Mysql.DBuser, config.Mysql.DBpass, config.Mysql.DBname)

	// TODO: send dump to remote receiver

	// TODO: if remote receiver, get dump to restore

	// restore DB
	database.RestoreDatabase(db, mysqlDump)
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
