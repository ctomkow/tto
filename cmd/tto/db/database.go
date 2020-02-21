// 2019 Craig Tomkow

// the generic database file that defines the interface
package db

import (
	"bufio"
	"github.com/ctomkow/tto/cmd/tto/exec"
	"io"
)

type DB interface {
	// open connection to database
	Open() error

	// create database
	Create() error

	// drop database
	Drop() error

	// dump the database with the command line utility
	Dump(exe *exec.Exec) (*io.ReadCloser, error)

	// restore the database using the database driver
	Restore(reader *bufio.Reader) error

	// return the implementation type
	Impl() string

	// return the name of the database
	Name() string

	// return the filename of the dump
	DumpName() string
}
