// Craig Tomkow
// July 24, 2019

package db

import (
	"bufio"
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Database struct {

	// the opened db connection
	connection *sql.DB

	// type of database, mysql, postgres, etc
	impl string

	ip       net.IPAddr
	port     uint16
	username string
	password string
	name     string

	maxConnections int
}

func (db *Database) Make(dbImpl string, ip net.IPAddr, port uint16, username string, password string, dbName string, maxConn int) {

	db.impl = dbImpl
	db.ip = ip
	db.port = port
	db.username = username
	db.password = password
	db.name = dbName
	db.maxConnections = maxConn
}

func (db *Database) Open() error {

	var conn *sql.DB
	var err error

	if strings.Compare(db.impl, "mysql") == 0 {
		conn, err = sql.Open(db.impl, db.username+":"+db.password+"@tcp("+db.ip.String()+":"+strconv.FormatUint(uint64(db.port), 10)+")/"+db.name)
		if err != nil {
			return err
		}
	} else {
		return errors.New("unsupported database type")
	}

	// set a static max number of concurrent connections allowed in the pool. 10 seems like a good number (see link), definitely not unlimited like the default!
	// https://www.alexedwards.net/blog/configuring-sqldb
	conn.SetMaxOpenConns(db.maxConnections)

	err = conn.Ping()
	if err != nil {
		return err
	}

	db.connection = conn

	return nil
}

func (db *Database) Dump(workingDir string) ([]byte, string, error) {

	// YYYYMMDDhhmmss
	currentTime := time.Now().UTC().Format("20060102150405") //TODO: remove static time format (or move it), buffer also relies on this format

	ipArg := "-h" + db.ip.String()
	portArg := "-P" + strconv.FormatUint(uint64(db.port), 10)
	userArg := "-u" + db.username
	passArg := "-p" + db.password
	sqlFile := db.name + "-" + currentTime + ".sql"
	var cmd *exec.Cmd

	if strings.Compare(db.impl, "mysql") == 0 {
		cmd = exec.Command("mysqldump", "--single-transaction", "--skip-lock-tables", "--routines", "--triggers", ipArg, portArg, userArg, passArg, db.name)
	} else {
		return nil, "", errors.New("unsupported database type")
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, "", err
	}

	if err = cmd.Start(); err != nil {
		return nil, "", err
	}

	byteBuffer, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, "", err
	}

	if err = cmd.Wait(); err != nil {
		return nil, "", err
	}

	return byteBuffer, sqlFile, nil
}

func (db *Database) Restore(dump string) error {

	// read .sql statement by statement and fire off to database server
	// NOTE: bufio.NewScanner has a line length limit of 65536 chars. mysqldump does only one INSERT per table. Not good!
	//       Using ReadString with a ';' delimiter, ensuring that the next character after is '\n'
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

func (db *Database) Drop() error {

	_, err := db.connection.Exec("DROP DATABASE " + db.name + ";")
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) Create() error {

	_, err := db.connection.Exec("CREATE DATABASE " + db.name + ";")
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) GetName() string {

	return db.name
}
