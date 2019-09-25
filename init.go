// Craig Tomkow
// August 6, 2019

package main

import (
	"encoding/json"
	"github.com/ctomkow/tto/conf"
	"github.com/golang/glog"
	"os"
	"os/user"
	"strconv"
)

func install() {

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

		// populate with sample configuration
		var sampleConf = new(conf.Config)
		sampleConf.MakeConfig()

		var jsonData []byte
		jsonData, err = json.MarshalIndent(sampleConf, "", "    ")
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

func setupWorkingDir(conf *conf.Config) {

	// ensure working directory files exists
	if !fileExists(conf.System.WorkingDir + ".latest.dump") {
		_, err := os.Create(conf.System.WorkingDir + ".latest.dump")
		if err != nil {
			glog.Exit(err)
		}
		glog.Info("created file: " + conf.System.WorkingDir + ".latest.dump")
	}

	// ensure working directory files exists
	if !fileExists(conf.System.WorkingDir + ".latest.restore") {
		_, err := os.Create(conf.System.WorkingDir + ".latest.restore")
		if err != nil {
			glog.Exit(err)
		}
		glog.Info("created file: " + conf.System.WorkingDir + ".latest.restore")
	}
}

func setupPermissions(conf *conf.Config) {

	// chown all files to appropriate usr

	// get app uid/gid based on system.conf from conf.json
	usr, err := user.Lookup(conf.System.User)
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
}

func fileExists(filename string) bool {

	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
