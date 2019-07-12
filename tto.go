package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)
import "fmt"

type config struct {
	System struct {
		Role string   `json:"role"`
		Replicate struct {
			Mysql string `json:"mysql"`
		}
	}
	Mysql struct {
		DBip   string `json:"db_ip"`
		DBport string `json:"db_port"`
		DBname string `json:"db_name"`
		DBuser string `json:"db_user"`
		DBpass string `json:"db_pass"`
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

	// parse config file input
	config, _ := loadConfig(*configFile)



	fmt.Println(config.System.Replicate.Mysql)
}

// parse -conf flag and return as pointer
func cliFlags() *string {

	confFilePtr := flag.String("conf", "conf.json", "name of conf file. default name is conf.json")
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

func dumpDatabase(dbIP string, dbPort string, )