// Craig Tomkow
// August 2, 2019

package main

import (
	"errors"
	"github.com/ctomkow/tto/configuration"
	"github.com/ctomkow/tto/database"
	"github.com/ctomkow/tto/processes"
	"github.com/ctomkow/tto/remote"
	"github.com/golang/glog"
	"github.com/robfig/cron"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func Sender(conf *configuration.Config) error {

	// setup various components
	//   - signal interrupts
	//   - local database connection
	//   - ring buffer for tracking database dumps
	//   - ssh connection to remote host
	//   - cron scheduling
	//   - ticker to check on ssh connection

	interrupt := SetupSignal()
	db := SetupDatabase(conf)
	buff := setupBuffer(conf.System.Role.Sender.MaxBackups)
	remoteHost := setupSSH(conf)
	cronChannel, cronjob := setupCron(conf.System.Role.Sender.Cron)
	ticker, testSSH := setupTicker(60)

	// database dump prep and manipulation
	//   - get the existing database dumps
	//   - parse and sort them for only the timestamps
	//   - add timestamps to ring buffer
	//   - delete remote database dumps that didn't fit into ring buffer
	//   - start ticker that monitors ssh connection

	if err := remoteHost.Connect(); err != nil {
		return err
	}
	remoteAlive := true
	remoteDBdumps, err := processes.GetRemoteDumps(remoteHost, conf.System.Role.Sender.DBname, conf.System.WorkingDir)
	if err != nil {
		return err
	}
	sortedDbDumpTimestamps := ParseDbDumpFilename(remoteDBdumps)
	buffOverflowTimestamps := fillBuffer(buff, conf.System.Role.Sender.DBname, sortedDbDumpTimestamps)
	buffOverflowDbDumpNames := buildDbDumpNames(conf.System.Role.Sender.DBname, buffOverflowTimestamps)
	if err := processes.DeleteRemoteDumps(remoteHost, conf.System.WorkingDir, buffOverflowDbDumpNames); err != nil {
		glog.Error(err)
	}

	cronjob.Start()
	startTicker(ticker, testSSH)

	for {
		select {

		// test ssh connection
		case <-testSSH:

			err := remoteHost.NewSession()

			if err == nil {
				if err = remoteHost.CloseSession(); err != nil {
					glog.Error("could not close test ssh session")
				}
				break
			}
			glog.Error("remote connection is down. backups are suspended until connection is re-established")

			// try re-connecting 3 times with a sleep of 1 minutes in-between
			for i := 1; i <= 3; i++ {
				glog.Error("[" + strconv.Itoa(i) + "/3]" + " attempting to re-connect with remote")
				if err := remoteHost.Connect(); err != nil {
					glog.Error("failed to re-establish connection with remote")
				} else {
					remoteAlive = true
					glog.Info("re-established connection with remote")
					break // success!
				}
				remoteAlive = false
				time.Sleep(10 * time.Second)
			}

		// cron trigger
		case <-cronChannel:

			if !remoteAlive {
				glog.Error("remote is down")
				break
			}
			mysqlDump, err := db.Dump(conf.System.WorkingDir)
			if err != nil {
				glog.Error(err)
				break
			}

			err = processes.TransferDumpToRemote(remoteHost, conf.System.WorkingDir, mysqlDump)
			if err != nil {
				glog.Error(err)
				break
			}

			buffOverflowTimestamp := buff.Enqueue(conf.System.Role.Sender.DBname, ParseDbDumpFilename(mysqlDump)[0])
			if buffOverflowTimestamp.IsZero() {
				break
			}

			// delete the dump that get's kicked out of the ring buffer
			var buffOverflowFilenames []string
			buffOverflowFilenames = append(buffOverflowFilenames, CompileDbDumpFilename(conf.System.Role.Sender.DBname, buffOverflowTimestamp))
			if err := processes.DeleteRemoteDumps(remoteHost, conf.System.WorkingDir, buffOverflowFilenames); err != nil {
				glog.Error(err)
				break
			}

		// trigger on signal
		case killSignal := <-interrupt:

			glog.Error(killSignal)

			if killSignal == os.Interrupt {
				return errors.New("daemon was interrupted by system signal")
			}
			return errors.New("daemon was killed")
		}
	}

	return nil
}

func cronTriggered(c chan bool) {

	c <- true
}

func SetupSignal() chan os.Signal {

	// Setup channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	return interrupt
}

func SetupDatabase(conf *configuration.Config) *database.Database {

	// setup database connection for sender
	// default max db connections is 10
	var db = new(database.Database)
	db.Make(conf.System.Role.Sender.Database, conf.System.Role.Sender.DBip, conf.System.Role.Sender.DBport,
		conf.System.Role.Sender.DBuser, conf.System.Role.Sender.DBpass, conf.System.Role.Sender.DBname, 10)

	return db
}

func setupSSH(conf *configuration.Config) *remote.SSH {

	// setup remote SSH connection
	var remoteConnPtr = new(remote.SSH)
	remoteConnPtr.Make(conf.System.Role.Sender.Dest.String(), strconv.FormatUint(uint64(conf.System.Role.Sender.Port), 10),
		conf.System.User, conf.System.Pass)

	glog.Info("receiver host: " + conf.System.Role.Sender.Dest.String())
	return remoteConnPtr
}

func setupBuffer(maxSize int) *CircularQueue {

	var buff = new(CircularQueue)
	buff.Make(maxSize)

	glog.Info("maximum backups: " + strconv.Itoa(maxSize))
	return buff
}

func fillBuffer(buff *CircularQueue, databaseName string, sortedTimestamps []time.Time) []time.Time {

	buffOverflowTimestamps := buff.Populate(databaseName, sortedTimestamps)

	for _, elem := range buff.queue[0:buff.size] {
		glog.Info("existing backups: " + elem.name)
	}

	return buffOverflowTimestamps
}

func buildDbDumpNames(databaseName string, times []time.Time) []string {

	// convert array of time.Time into array of DB dump filenames
	var dbDumpNames []string
	for _, timestamp := range times {
		dbDumpNames = append(dbDumpNames, CompileDbDumpFilename(databaseName, timestamp))
	}

	return dbDumpNames
}

func setupCron(cronStatement string) (chan bool, *cron.Cron) {

	// cron setup
	cronChannel := make(chan bool)
	cj := cron.New()
	cj.AddFunc(cronStatement, func() { cronTriggered(cronChannel) })

	glog.Info("db backup schedule: " + cronStatement)
	return cronChannel, cj
}

func setupTicker(secInterval time.Duration) (*time.Ticker, chan bool) {

	ticker := time.NewTicker(secInterval * time.Second)
	tickChannel := make(chan bool)

	return ticker, tickChannel
}

func startTicker(ticker *time.Ticker, tickerChannel chan bool) {

	go func() {
		for _ = range ticker.C {
			tickerChannel <- true
		}
	}()
}
