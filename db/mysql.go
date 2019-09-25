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
	"net"
	"os"
	"strconv"
	"strings"
)

type Database struct {

	// the opened db Connection
	Connection *sql.DB

	// type of database, mysql, postgres, etc
	Impl string

	Ip       net.IPAddr
	Port     uint16
	Username string
	Password string
	Name     string

	maxConnections int
}

func (db *Database) Make(dbImpl string, ip net.IPAddr, port uint16, username string, password string, dbName string, maxConn int) {

	db.Impl = dbImpl
	db.Ip = ip
	db.Port = port
	db.Username = username
	db.Password = password
	db.Name = dbName
	db.maxConnections = maxConn
}

func (db *Database) Open() error {

	var conn *sql.DB
	var err error

	if strings.Compare(db.Impl, "mysql") == 0 {
		conn, err = sql.Open(db.Impl, db.Username+":"+db.Password+"@tcp("+db.Ip.String()+":"+strconv.FormatUint(uint64(db.Port), 10)+")/"+db.Name)
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

	db.Connection = conn

	return nil
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
			_, err = db.Connection.Exec(buffer.String())
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

	_, err := db.Connection.Exec("DROP DATABASE " + db.Name + ";")
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) Create() error {

	_, err := db.Connection.Exec("CREATE DATABASE " + db.Name + ";")
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) GetName() string {

	return db.Name
}
