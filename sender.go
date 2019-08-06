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
)

func Sender(conf *configuration.Config) error {

	// Setup channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// setup database connection for sender
	var db = new(database.Database)
	db.Make(conf.System.Role.Sender.Database, conf.System.Role.Sender.DBip, conf.System.Role.Sender.DBport,
		conf.System.Role.Sender.DBuser, conf.System.Role.Sender.DBpass, conf.System.Role.Sender.DBname)

	// setup remote SSH connection
	var remoteConnPtr = new(remote.SSH)
	remoteConnPtr.Make(conf.System.Role.Sender.Dest.String(), strconv.FormatUint(uint64(conf.System.Role.Sender.Port), 10),
		conf.System.User, conf.System.Pass)
	if err := remoteConnPtr.Connect(); err != nil {
		return err
	}

	// get remote files
	remoteFiles, err := processes.GetRemoteDumps(remoteConnPtr, conf.System.Role.Sender.DBname, conf.System.WorkingDir)
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
	glog.Info(errors.New("ring buffer filled upto max_backups (" + strconv.Itoa(numOfBackups) + ") with existing remote db dumps: "))
	if err != nil {
		glog.Fatal(err)
	}
	for _, elem := range buff.queue[0:buff.size] {
		glog.Info(errors.New(elem.name))
	}

	// convert array of time.Time into array of DB dump filenames
	var buffOverflowFilenames []string
	for _, elem := range buffOverflowTimestamps {
		buffOverflowFilenames = append(buffOverflowFilenames, CompileDbDumpFilename(conf.System.Role.Sender.DBname, elem))
	}

	// delete any remote files that don't fit into ring buffer
	if err := processes.DeleteRemoteDump(remoteConnPtr, conf.System.WorkingDir, buffOverflowFilenames); err != nil {
		glog.Error(err)
	}
	glog.Info(errors.New("the following remote db dumps that didn't fit in ring buffer were deleted: "))
	for _, elem := range buffOverflowFilenames {
		glog.Info(errors.New(elem))
	}

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
				copiedDump, err := processes.TransferDumpToRemote(remoteConnPtr, conf.System.WorkingDir, mysqlDump)
				if err != nil {
					glog.Error(err)
				}
				glog.Info(errors.New("copied over db: " + copiedDump))

				// add to ring buffer and delete any overwritten file
				buffOverflowTimestamp := buff.Enqueue(conf.System.Role.Sender.DBname, ParseDbDumpFilename(mysqlDump)[0])
				if !buffOverflowTimestamp.IsZero() {

					// convert array of time.Time into array of DB dump filenames
					var buffOverflowFilenames []string
					buffOverflowFilenames = append(buffOverflowFilenames, CompileDbDumpFilename(conf.System.Role.Sender.DBname, buffOverflowTimestamp))

					if err := processes.DeleteRemoteDump(remoteConnPtr, conf.System.WorkingDir, buffOverflowFilenames); err != nil {
						glog.Error(err)
					}
					glog.Info(errors.New("deleted old db: " + CompileDbDumpFilename(conf.System.Role.Sender.DBname, buffOverflowTimestamp)))
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

func cronTriggered(c chan bool) {

	c <- true
}
