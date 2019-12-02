// 2019 Craig Tomkow

package db

import (
	"bufio"
	"database/sql"
	"github.com/ctomkow/tto/exec"
	"github.com/ctomkow/tto/util"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"net"
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
	cmd      *exec.Exec
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
func (db *Mysql) Dump(exe *exec.Exec) (*io.ReadCloser, error) {
	db.filename = db.name + "_-_" + util.NewTimestamp().Timestamp() + ".sql"

	ipArg := "-h" + db.ip.String()
	portArg := "-P" + strconv.FormatUint(uint64(db.port), 10)
	userArg := "-u" + db.user
	passArg := "-p" + db.pass

	exe.LocalCmdOnly([]string{"mysqldump", "--single-transaction", "--skip-lock-tables", "--routines", "--triggers", ipArg, portArg, userArg, passArg, db.name})

	stdout, err := exe.Cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err = exe.Cmd.Start(); err != nil {
		return nil, err
	}

	return &stdout, nil
}

// Read database dump statement by statement and fire off to the database
// Note: bufio.NewScanner has a line length limit of 65536 chars. db dump does only one INSERT per table
// Using ReadString with a ';' delimiter, ensuring that the next character after is '\n'
func (db *Mysql) Restore(reader *bufio.Reader) error {
	var buf strings.Builder
	for {
		statement, err := reader.ReadString(';')
		if err != nil {
			if err == io.EOF {
				break
			}
		}

		buf.WriteString(statement)

		// look at next byte
		nextByte, err := reader.Peek(1)
		if err != nil {
			return err
		}

		// newline '\n' aka utf decimal '10'
		if nextByte[0] == 10 {
			_, err = db.connection.Exec(buf.String())
			if err != nil {
				return err
			}
			buf.Reset()
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

// return dump filename
func (db *Mysql) DumpName() string {
	return db.filename
}
