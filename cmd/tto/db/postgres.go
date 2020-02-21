// 2019 Craig Tomkow

// the database actions for postgres
package db

import (
	"database/sql"
	"net"
	"os/exec"
)

type Postgres struct {

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

	// used for postgresdump
	cmd      *exec.Cmd
	filename string
}

// instantiate a new postgres struct
func NewPostgres() {}

// connect to database and ensure it is reachable
func (db *Postgres) Open() {}

// create the database
func (db *Postgres) Create() {}

// drop the database
func (db *Postgres) Drop() {}

// dump the database and return the stdout stream
func (db *Postgres) Dump() {}

// restore database
func (db *Postgres) Restore() {}

// return implementation type
func (db *Postgres) Impl() {}

// return database name
func (db *Postgres) Name() {}
