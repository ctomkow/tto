// Craig Tomkow
// July 24, 2019

package database

import (
	"bufio"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ##### public functions #####

func Open(dbPort string, dbIp string, dbUser string, dbPass string, dbName string) (*sql.DB, error) {

	// prep DB connection
	db, err := sql.Open("mysql", dbUser+":"+dbPass+"@tcp("+dbIp+":"+dbPort+")/"+dbName)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Dump(dbPort string, dbIp string, dbUser string, dbPass string, dbName string, workingDir string) (string, error) {

	// YYYYMMDDhhmmss
	currentTime := time.Now().UTC().Format("20060102150405") //TODO: remove static time format (or move it), buffer also relies on this format

	portArg := "-P" + dbPort
	ipArg := "-h" + dbIp
	userArg := "-u" + dbUser
	passArg := "-p" + dbPass
	sqlFile := dbName + "-" + currentTime + ".sql"

	cmd := exec.Command("mysqldump", "--single-transaction", "--routines", "--triggers", portArg, ipArg, userArg, passArg, dbName)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	defer func() {
		if err := stdout.Close(); err != nil {
			glog.Error(err)
		}
	}()

	if err = cmd.Start(); err != nil {
		return "", err
	}

	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", err
	}

	if err = ioutil.WriteFile(workingDir+sqlFile, bytes, 0644); err != nil {
		return "", err
	}

	return sqlFile, nil
}

func Restore(db *sql.DB, dump string) error {

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

	// loop and send queries until EOF
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

		//fmt.Print(buffer.String())
		//fmt.Printf("%q", oracleBytes)

		if oracleBytes[0] == 10 { // newline '\n' aka utf decimal '10'
			_, err = db.Exec(buffer.String())
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

func Drop(db *sql.DB, dbName string) error {

	_, err := db.Exec("DROP DATABASE " + dbName + ";")
	if err != nil {
		return err
	}

	return nil
}

func Create(db *sql.DB, dbName string) error {

	_, err := db.Exec("CREATE DATABASE " + dbName + ";")
	if err != nil {
		return err
	}

	return nil
}
