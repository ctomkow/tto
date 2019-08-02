// Craig Tomkow
// August 2, 2019

package configuration

import (
	"encoding/json"
	"github.com/golang/glog"
	"net"
	"os"
)

type Config struct {
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

func (conf *Config) MakeConfig() {

	conf.System.User = `username`
	conf.System.Pass = `password`
	conf.System.WorkingDir = `/opt/tto/`
	conf.System.Type = `sender|receiver`
	conf.System.Role.Sender.Dest = net.IPAddr{net.IPv4(6, 6, 6, 6), ""}
	conf.System.Role.Sender.Port = uint16(22)
	conf.System.Role.Sender.Database = `mysql`
	conf.System.Role.Sender.DBip = net.IPAddr{net.IPv4(7, 7, 7, 7), ""}
	conf.System.Role.Sender.DBport = uint16(3306)
	conf.System.Role.Sender.DBuser = `username`
	conf.System.Role.Sender.DBpass = `password`
	conf.System.Role.Sender.DBname = `databaseName`
	conf.System.Role.Sender.Cron = `a cron statement`
	conf.System.Role.Sender.MaxBackups = int(5)
	conf.System.Role.Receiver.Database = `mysql`
	conf.System.Role.Receiver.DBip = net.IPAddr{net.IPv4(8, 8, 8, 8), ""}
	conf.System.Role.Receiver.DBport = uint16(3306)
	conf.System.Role.Receiver.DBuser = `username`
	conf.System.Role.Receiver.DBpass = `password`
	conf.System.Role.Receiver.DBname = `databaseName`
}

func (conf *Config) LoadConfig(filename string) error {

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
