// Craig Tomkow
// August 2, 2019

package main

import (
	"errors"
	"github.com/ctomkow/tto/configuration"
	"github.com/ctomkow/tto/database"
	"github.com/ctomkow/tto/processes"
	"github.com/golang/glog"
	"github.com/robfig/cron"
	"os"
	"os/signal"
	"syscall"
)

func Sender(conf *configuration.Config) error {

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// setup database connection for sender
	var db = new(database.Database)
	db.Make(conf.System.Role.Sender.Database, conf.System.Role.Sender.DBip, conf.System.Role.Sender.DBport,
		conf.System.Role.Sender.DBuser, conf.System.Role.Sender.DBpass, conf.System.Role.Sender.DBname)

	// get remote files
	remoteFiles, err := processes.GetRemoteDumps(conf.System.Role.Sender.Dest, conf.System.Role.Sender.Port,
		conf.System.User, conf.System.Pass, conf.System.Role.Sender.DBname, conf.System.WorkingDir)
	if err != nil {
		glog.Fatal(err)
	}

	// init ring buffer with existing files
	var buff = new(CircularQueue)
	sortedTimeSlice := ParseDbDumpFilename(remoteFiles)
	numOfBackups := conf.System.Role.Sender.MaxBackups
	if err != nil {
		glog.Fatal(err)
	}
	buffOverflowTimestamps := buff.Make(numOfBackups, conf.System.Role.Sender.DBname, sortedTimeSlice)
	glog.Info(errors.New("ring buffer filled up to max_backups with existing database dumps"))

	// convert array of time.Time into array of DB dump filenames
	var buffOverflowFilenames []string
	for _, elem := range buffOverflowTimestamps {
		buffOverflowFilenames = append(buffOverflowFilenames, CompileDbDumpFilename(conf.System.Role.Sender.DBname, elem))
	}

	// delete any remote files that don't fit into ring buffer
	if err := processes.DeleteRemoteDump(conf.System.Role.Sender.Dest, conf.System.Role.Sender.Port, conf.System.User,
		conf.System.Pass, conf.System.Role.Sender.DBname, conf.System.WorkingDir, buffOverflowFilenames); err != nil {
		glog.Error(err)
	}
	glog.Info(errors.New("database dumps on remote machine that didn't fit in ring buffer have been deleted"))

	// cron setup
	cronChannel := make(chan bool)
	cj := cron.New()
	err = cj.AddFunc(conf.System.Role.Sender.Cron, func() { cronTriggered(cronChannel) })
	if err != nil {
		glog.Fatal(err)
	}
	cj.Start()

	for {
		select {

		// cron trigger
		case <-cronChannel:
			mysqlDump, err := db.Dump(conf.System.WorkingDir)
			if err != nil {
				glog.Error(err)
			}
			if err == nil {
				copiedDump, err := processes.TransferDumpToRemote(conf.System.Role.Sender.Dest, conf.System.Role.Sender.Port,
					conf.System.User, conf.System.Pass, conf.System.WorkingDir, mysqlDump)
				if err != nil {
					glog.Error(err)
				}
				glog.Info(errors.New("dumped and copied over database: " + copiedDump))

				// add to ring buffer and delete any overwritten file
				buffOverflowTimestamp := buff.Enqueue(conf.System.Role.Sender.DBname, ParseDbDumpFilename(mysqlDump)[0])
				if !buffOverflowTimestamp.IsZero() {

					// convert array of time.Time into array of DB dump filenames
					var buffOverflowFilenames []string
					buffOverflowFilenames = append(buffOverflowFilenames, CompileDbDumpFilename(conf.System.Role.Sender.DBname, buffOverflowTimestamp))

					if err := processes.DeleteRemoteDump(conf.System.Role.Sender.Dest, conf.System.Role.Sender.Port,
						conf.System.User, conf.System.Pass, conf.System.Role.Sender.DBname, conf.System.WorkingDir,
						buffOverflowFilenames); err != nil {
						glog.Error(err)
					}
					glog.Info(errors.New("deleted old database dump: " + CompileDbDumpFilename(conf.System.Role.Sender.DBname, buffOverflowTimestamp)))
				}
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
