// 2019 Craig Tomkow

package db

import (
	"bufio"
	"database/sql"
	"github.com/ctomkow/tto/util"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Mysql struct {

	// the opened db Connection
	connection *sql.DB

	// type of database, mysql, postgres, etc
	impl string

	// database connection details
	ip      net.IPAddr
	port    uint16
	user    string
	pass    string
	name    string
	maxConn int

	// used for mysqldump
	cmd      *exec.Cmd
	filename string
}

// instantiate a new mysql struct
func NewMysql(impl string, ip net.IPAddr, port uint16, user string, pass string, name string, maxConn int) *Mysql {

	return &Mysql{
		connection: nil,
		impl:       impl,
		ip:         ip,
		port:       port,
		user:       user,
		pass:       pass,
		name:       name,
		maxConn:    maxConn,
		cmd:        nil,
	}
}

// connect to database and ensure it is reachable
func (db *Mysql) Open() error {
	conn, err := sql.Open(db.impl, db.user+":"+db.pass+"@tcp("+db.ip.String()+":"+strconv.FormatUint(uint64(db.port), 10)+")/"+db.name)
	if err != nil {
		return err
	}

	conn.SetMaxOpenConns(db.maxConn)

	err = conn.Ping()
	if err != nil {
		return err
	}

	db.connection = conn

	return nil
}

// create the database
func (db *Mysql) Create() error {
	_, err := db.connection.Exec("CREATE DATABASE " + db.name + ";")
	if err != nil {
		return err
	}

	return nil
}

// drop the database
func (db *Mysql) Drop() error {
	_, err := db.connection.Exec("DROP DATABASE " + db.name + ";")
	if err != nil {
		return err
	}

	return nil
}

// dump the database and return the stdout stream
func (db *Mysql) Dump() (*io.ReadCloser, error) {
	db.filename = db.name + "-" + util.NewTimestamp().Timestamp() + ".sql"

	ipArg := "-h" + db.ip.String()
	portArg := "-P" + strconv.FormatUint(uint64(db.port), 10)
	userArg := "-u" + db.user
	passArg := "-p" + db.pass

	db.cmd = exec.Command("mysqldump", "--single-transaction", "--skip-lock-tables", "--routines", "--triggers", ipArg, portArg, userArg, passArg, db.name)

	stdout, err := db.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err = db.cmd.Start(); err != nil {
		return nil, err
	}

	return &stdout, nil
}

// TODO: refactor
// read .sql statement by statement and fire off to database server
// NOTE: bufio.NewScanner has a line length limit of 65536 chars. mysqldump does only one INSERT per table. Not good!
//       Using ReadString with a ';' delimiter, ensuring that the next character after is '\n'
func (db *Mysql) Restore(dump string) error {

	fd, err := os.Open(dump)
	if err != nil {
		return err
	}
	defer func() {
		if err := fd.Close(); err != nil {
			glog.Error(err)
		}
	}()

	reader := bufio.NewReader(fd)
	var buffer strings.Builder

	// send queries until EOF
	for {

		statement, err := reader.ReadString(';')
		if err != nil {
			if err == io.EOF {
				break
			}
		}

		buffer.WriteString(statement)

		// look at next byte
		oracleBytes, err := reader.Peek(1)
		if err != nil {
			return err
		}

		if oracleBytes[0] == 10 { // newline '\n' aka utf decimal '10'
			_, err = db.connection.Exec(buffer.String())
			if err != nil {
				return err
			}
			buffer.Reset()
		} else {
			continue
		}
	}

	return nil
}

// return implementation type
func (db *Mysql) Impl() string {
	return db.impl
}

// return database name
func (db *Mysql) Name() string {
	return db.name
}
