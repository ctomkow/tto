// Craig Tomkow
// August 2, 2019

package main

import (
	"errors"
	"github.com/ctomkow/tto/cmd/tto/backup"
	"github.com/ctomkow/tto/cmd/tto/conf"
	"github.com/ctomkow/tto/cmd/tto/db"
	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	"net"
	"os"
	"time"
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
	//   - os exec process handling

	interrupt := newSignal()
	dB := newReceiverDb(
		conf.System.Role.Receiver.Database,
		conf.System.Role.Receiver.DBip,
		conf.System.Role.Receiver.DBport,
		conf.System.Role.Receiver.DBuser,
		conf.System.Role.Receiver.DBpass,
		conf.System.Role.Receiver.DBname,
		10,
	)
	watcher, err := newFileWatcher()
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
	exe := newExecHandler()

	// create working components
	//   - open database connection
	//   - watch file for changes
	//   - file change event variable

	// useful for when the tto server boots up faster than the database server
	if err := attemptDB(dB, 3, 10); err != nil {
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
			output, err := exe.LocalCmd(conf.System.Role.Receiver.ExecBefore)
			if err != nil {
				glog.Error(err)
				lck.restore = false
				break
			}
			glog.Info(errors.New(output))

			// run restoreDatabase as a goroutine. goroutine holds a restoreDatabase lock until it's done
			go func() {
				restoredDump, err := backup.Restore(dB, conf.System.WorkingDir)
				if err != nil {
					glog.Error(err)
					restoreChan <- ""
					return
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
			output, err := exe.LocalCmd(conf.System.Role.Receiver.ExecAfter)
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

func newFileWatcher() (*fsnotify.Watcher, error) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return watcher, err
	}

	return watcher, nil
}

// factory to setup chosen database
func newReceiverDb(impl string, ip net.IPAddr, port uint16, user string, pass string, name string, maxConn int) db.DB {
	switch impl {
	case "mysql":
		return db.NewMysql(impl, ip, port, user, pass, name, maxConn)
	case "postgres":
		// pass
	default:
		return nil
	}
	return nil
}

func attemptDB(dB db.DB, tries int, delayInSec int) error {
	var err error
	for i := 1; i <= tries; i++ {
		if err = dB.Open(); err == nil {
			return nil
		}
		time.Sleep(time.Duration(delayInSec) * time.Second)
	}
	return err
}
