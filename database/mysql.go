package database

import (
	"bufio"
	"database/sql"
	"strings"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
	_ "github.com/go-sql-driver/mysql"
)

func connectToDatabase (dbPort string, dbIp string, dbUser string, dbPass string, dbName string) (db *sql.DB) {

	// prep DB connection
	db, err := sql.Open("mysql", dbUser + ":" + dbPass + "@tcp(" + dbIp + ":" + dbPort + ")/" + dbName)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()

	if err != nil {
		log.Fatal(err)
	}

	return db
}


func dropDatabase (db *sql.DB, dbName string) {

	_, err := db.Exec("DROP DATABASE " + dbName + ";")
	if err != nil {
		log.Fatal(err)
	}
}

func createDatabase (db *sql.DB, dbName string) {

	_, err := db.Exec("CREATE DATABASE " + dbName + ";")
	if err != nil {
		log.Fatal(err)
	}

}

func dumpDatabase(dbPort string, dbIp string, dbUser string, dbPass string, dbName string) string {

	// YYYYMMDDhhmmss
	currentTime := time.Now().Format("20060102150405")

	portArg := "-P" + dbPort
	ipArg   := "-h" + dbIp
	userArg := "-u" + dbUser
	passArg := "-p" + dbPass
	sqlFile := dbName + currentTime + ".sql"

	cmd := exec.Command("mysqldump", "--single-transaction", "--routines", "--triggers", portArg, ipArg, userArg, passArg, dbName)
	stdout, err := cmd.StdoutPipe()
	defer stdout.Close()

	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Start()

	if err != nil {
		log.Fatal(err)
	}

	bytes, err := ioutil.ReadAll(stdout)

	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(sqlFile, bytes, 0644)
	return sqlFile
}

func restoreDatabase(db *sql.DB, dump string) {

	// read .sql statement by statement and fire off to database server
	// NOTE: bufio.NewScanner has a line length limit of 65536 chars. mysqldump does only one INSERT per table. Not good!
	//       Using ReadString with a ';' delimiter, ensuring that the next character after is '\n'
	fd, err := os.Open(dump)
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()

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
		oracleByte, err := reader.Peek(1)
		if err != nil {
			log.Fatal(err)
		}

		//fmt.Print(buffer.String())
		//fmt.Printf("%q", oracleByte)

		// if next byte is '\n', then sql statement looks complete
		if oracleByte[0] == 10 {
			_, err = db.Exec(buffer.String())
			if err != nil {
				log.Fatal(err)
			}
			buffer.Reset()
		} else {
			continue
		}
	}

}