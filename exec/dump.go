// 2019 Craig Tomkow

package exec

import (
	"github.com/ctomkow/tto/util"
	"io"
	"os/exec"
)

type DB interface {
	// dump the database with the command line utility
	Dump(ip string, port string, user string, pass string, name string) (*io.ReadCloser, error)

	// return the filename of the dump
	Name() string
}

type MysqlDump struct {
	Cmd  *exec.Cmd
	name string
}

type PostgresDump struct {
	Cmd  *exec.Cmd
	name string
}

func (m *MysqlDump) Dump(ip string, port string, user string, pass string, name string) (*io.ReadCloser, error) {
	timestamp := util.NewTimestamp().Timestamp()
	ipArg     := "-h" + ip
	portArg   := "-P" + port
	userArg   := "-u" + user
	passArg   := "-p" + pass
	m.name     = name + "-" + timestamp + ".sql"

	m.Cmd = exec.Command("mysqldump", "--single-transaction", "--skip-lock-tables", "--routines", "--triggers", ipArg, portArg, userArg, passArg, name)

	stdout, err := m.Cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err = m.Cmd.Start(); err != nil {
		return nil, err
	}

	return &stdout, nil
}

func (m *PostgresDump) Dump(ip string, port string, user string, pass string, name string) (*io.ReadCloser, error) {
	// placeholder
	// TODO: implement postgresdump

	return nil, nil
}

func (m *MysqlDump) Name() string {
	return m.name
}

func (m *PostgresDump) Name() string {
	return m.name
}