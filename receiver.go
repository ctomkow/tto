// Craig Tomkow
// August 2, 2019

package main

import (
	"errors"
	"github.com/ctomkow/tto/configuration"
	"github.com/ctomkow/tto/database"
	"github.com/ctomkow/tto/local"
	"github.com/ctomkow/tto/processes"
	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	"os"
	"os/signal"
	"syscall"
)

type lock struct {
	restore bool
}

func Receiver(conf *configuration.Config) error {

	// Setup channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// setup database connection for receiver
	// default max db connections is 10
	var db = new(database.Database)
	db.Make(conf.System.Role.Receiver.Database, conf.System.Role.Receiver.DBip, conf.System.Role.Receiver.DBport,
		conf.System.Role.Receiver.DBuser, conf.System.Role.Receiver.DBpass, conf.System.Role.Receiver.DBname, 10)
	if err := db.Open(); err != nil {
		return err
	}

	// setup file watcher monitoring .latest.dump
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		glog.Fatal(err)
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			glog.Exit(err)
		}
	}()

	var lck = new(lock)

	// FYI, VIM doesn't create a WRITE event, only RENAME, CHMOD, REMOVE (then breaks future watching)
	// https://github.com/fsnotify/fsnotify/issues/94#issuecomment-287456396
	if err = watcher.Add(conf.System.WorkingDir + ".latest.dump"); err != nil {
		glog.Fatal(err)
	}
	var event fsnotify.Event

	// create channel for communicating with the database restoreDatabase routine
	restoreChan := make(chan string)

	for {
		select {

		// trigger on write event
		case event = <-watcher.Events:
			if !isWriteEvent(event) {
				break
			}
			if lck.restore {
				break
			}

			lck.restore = true

			// run exec_before
			output, err := local.RunCommand(conf.System.Role.Receiver.ExecBefore)
			if err != nil {
				glog.Error(err)
				lck.restore = false
				break
			}
			glog.Info(errors.New(output))

			// run restoreDatabase as a goroutine. goroutine holds a restoreDatabase lock until it's done
			go func() {
				restoredDump, err := processes.RestoreDatabase(db, conf.System.WorkingDir)
				if err != nil {
					glog.Error(err)
					restoreChan <- ""
				}
				restoreChan <- restoredDump
			}()

		// trigger on dump restoreDatabase being finished
		case restoredDump := <-restoreChan:

			if restoredDump == "" {
				glog.Error(errors.New("failed to restore database"))
			} else {
				glog.Info(errors.New("restored database: " + restoredDump))
			}

			// run exec_after
			output, err := local.RunCommand(conf.System.Role.Receiver.ExecAfter)
			if err != nil {
				glog.Error(err)
			} else {
				glog.Info(output)
			}

			lck.restore = false

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

func isWriteEvent(event fsnotify.Event) bool {

	if event.Op&fsnotify.Write == fsnotify.Write {
		return true
	}

	return false
}
