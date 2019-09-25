// Craig Tomkow
// August 2, 2019

package main

import (
	"errors"
	"github.com/ctomkow/tto/conf"
	"github.com/ctomkow/tto/db"
	"github.com/ctomkow/tto/exec"
	"github.com/ctomkow/tto/backup"
	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	"os"
)

type lock struct {
	restore bool
}

func Receiver(conf *conf.Config) error {

	// setup various components
	//   - signal interrupts
	//   - local database
	//   - file watcher
	//   - restore lock
	//   - restore channel for the restore database routine

	interrupt := SetupSignal()
	db := setupReceiverDatabase(conf)
	watcher, err := setupFileWatcher()
	if err != nil {
		return err
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			glog.Exit(err)
		}
	}()
	var lck = new(lock)
	restoreChan := make(chan string)

	// create working components
	//   - open database connection
	//   - watch file for changes
	//   - file change event variable

	// TODO: try 3 times in failure
	if err := db.Open(); err != nil {
		return err
	}
	// FYI, VIM doesn't create a WRITE event, only RENAME, CHMOD, REMOVE (then breaks future watching). https://github.com/fsnotify/fsnotify/issues/94#issuecomment-287456396
	if err = watcher.Add(conf.System.WorkingDir + ".latest.dump"); err != nil {
		return err
	}
	var event fsnotify.Event

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
			output, err := exec.LocalCmd(conf.System.Role.Receiver.ExecBefore)
			if err != nil {
				glog.Error(err)
				lck.restore = false
				break
			}
			glog.Info(errors.New(output))

			// run restoreDatabase as a goroutine. goroutine holds a restoreDatabase lock until it's done
			go func() {
				restoredDump, err := backup.RestoreDb(db, conf.System.WorkingDir)
				if err != nil {
					glog.Error(err)
					restoreChan <- ""
				}
				restoreChan <- restoredDump
			}()

		// trigger on dump restoreDatabase being finished
		case restoredDump := <-restoreChan:

			if restoredDump == "" {
				glog.Error(errors.New("failed to restore db dump"))
			} else {
				glog.Info(errors.New("restored db dump: " + restoredDump))
			}

			// run exec_after
			output, err := exec.LocalCmd(conf.System.Role.Receiver.ExecAfter)
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

func setupFileWatcher() (*fsnotify.Watcher, error) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return watcher, err
	}

	return watcher, nil
}

func setupReceiverDatabase(conf *conf.Config) *db.Database {

	// setup database connection for sender
	// default max db connections is 10
	var db = new(db.Database)
	db.Make(conf.System.Role.Receiver.Database, conf.System.Role.Receiver.DBip, conf.System.Role.Receiver.DBport,
		conf.System.Role.Receiver.DBuser, conf.System.Role.Receiver.DBpass, conf.System.Role.Receiver.DBname, 10)

	return db
}
