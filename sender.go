// Craig Tomkow
// August 2, 2019

package main

import (
	"errors"
	"github.com/ctomkow/tto/backup"
	"github.com/ctomkow/tto/conf"
	"github.com/ctomkow/tto/db"
	"github.com/ctomkow/tto/exec"
	"github.com/ctomkow/tto/inet"
	"github.com/golang/glog"
	"github.com/robfig/cron"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func Sender(conf *conf.Config) error {

	// setup various components
	//   - signal interrupts
	//   - local database connection
	//   - ring buffer for tracking database dumps
	//   - ssh connection to remote host
	//   - cron scheduling
	//   - ticker to check on ssh connection
	//   - os exec process handling

	interrupt := newSignal()
	dB := newSenderDb(
		conf.System.Role.Sender.Database,
		conf.System.Role.Sender.DBip,
		conf.System.Role.Sender.DBport,
		conf.System.Role.Sender.DBuser,
		conf.System.Role.Sender.DBpass,
		conf.System.Role.Sender.DBname,
	)
	buf := newRingBuf(conf.System.Role.Sender.MaxBackups)
	remote := newSSH(
		conf.System.Role.Sender.Dest,
		conf.System.Role.Sender.Port,
		conf.System.User,
		conf.System.Pass,
		conf.System.SSHkey,
	)
	cronChan, cronJob := newCron(conf.System.Role.Sender.Cron)
	tickerChan, ticker := newTicker(60)
	exe := newExecHandler()

	// database dump prep and manipulation
	//   - get the existing backups
	//   - parse the multiline string into an array
	//   - strip path
	//   - add sorted backups to ring buffer
	//   - delete backups that didn't fit into ring buffer
	//   - start ticker that monitors ssh connection

	if err := remote.Connect(); err != nil {
		return err
	}
	remoteAlive := true
	multilineStringBackups, err := backup.Retrieve(remote, exe, conf.System.Role.Sender.DBname, conf.System.WorkingDir)
	if err != nil {
		return err
	}
	backups, err := parseMultilineString(multilineStringBackups)
	if err != nil {
		glog.Fatal(err)
	}
	backups = stripPath(conf.System.WorkingDir, backups)
	expiredDumps := fillBuf(buf, sortBackups(backups))
	if err := backup.Delete(remote, exe, conf.System.WorkingDir, expiredDumps); err != nil {
		glog.Error(err)
	}
	cronJob.Start()
	startTicker(ticker, tickerChan)

	for {
		select {
		// test ssh connection
		case <-tickerChan:
			if err = remote.TestConnection(); err != nil {
				glog.Error(err)
			} else {
				break
			}
			glog.Error("remote connection is down. backups are suspended until connection is re-established")
			if err := remote.Reconnect(3, 10); err != nil {
				glog.Error(err)
				remoteAlive = false
			} else {
				remoteAlive = true
			}

		// cron trigger
		case <-cronChan:
			if !remoteAlive {
				glog.Error("remote is down")
				break
			}

			dumpStdout, err := dB.Dump()
			if err != nil {
				glog.Error(err)
				break
			}

			err = backup.ToRemote(remote, conf.System.WorkingDir, dB.DumpName(), dumpStdout, exe)
			if err != nil {
				glog.Error(err)
				break
			}
			expiredDump := buf.Enqueue(dB.DumpName())
			if expiredDump == "" {
				break
			}
			if err := backup.Delete(remote, exe, conf.System.WorkingDir, []string{expiredDump}); err != nil {
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

// Setup channel on which to send signal notifications.
// We must use a buffered channel or risk missing the signal
// if we're not ready to receive when the signal is sent.
func newSignal() chan os.Signal {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	return interrupt
}

// factory to setup chosen database
func newSenderDb(impl string, ip net.IPAddr, port uint16, user string, pass string, name string) db.DB {
	switch impl {
	case "mysql":
		return db.NewMysql(impl, ip, port, user, pass, name, 0)
	case "postgres":
		// pass
	default:
		return nil
	}
	return nil
}

// create new ring buffer with a maximum size
func newRingBuf(size int) *CircularQueue {
	var buf = new(CircularQueue)
	buf.Make(size)
	glog.Info("maximum backups: " + strconv.Itoa(size))
	return buf
}

// setup new ssh connection with remote host
func newSSH(ip net.IPAddr, port uint16, user string, pass string, key string) *inet.SSH {
	var remoteConn = new(inet.SSH)
	remoteConn.Make(ip.String(), strconv.FormatUint(uint64(port), 10), user, pass, key)
	glog.Info("receiver host: " + ip.String())
	return remoteConn
}

// fill ring buffer with provided sorted backup names
func fillBuf(buf *CircularQueue, sortedBackups []string) []string {
	expiredBuffElements := buf.Populate(sortedBackups)
	for _, elem := range buf.queue[0:buf.size] {
		glog.Info("existing backups: " + elem.name)
	}
	return expiredBuffElements
}

// create a channel and cronjob
func newCron(schedule string) (chan bool, *cron.Cron) {
	channel := make(chan bool)
	cj := cron.New()
	cj.AddFunc(schedule, func() { cronTriggered(channel) })
	glog.Info("db backup schedule: " + schedule)
	return channel, cj
}

// create a channel and tick on every interval
func newTicker(secInterval time.Duration) (chan bool, *time.Ticker) {
	ticker := time.NewTicker(secInterval * time.Second)
	channel := make(chan bool)
	return channel, ticker
}

// start ticking
func startTicker(ticker *time.Ticker, channel chan bool) {
	go func() {
		for _ = range ticker.C {
			channel <- true
		}
	}()
}

// create new process exec handler
func newExecHandler() *exec.Exec {
	var exe = new(exec.Exec)
	return exe
}
