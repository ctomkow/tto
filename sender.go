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

	interrupt := SetupSignal()
	dB := setupSenderDB(
		conf.System.Role.Sender.Database,
		conf.System.Role.Sender.DBip,
		conf.System.Role.Sender.DBport,
		conf.System.Role.Sender.DBuser,
		conf.System.Role.Sender.DBpass,
		conf.System.Role.Sender.DBname,
	)
	buff := setupBuffer(conf.System.Role.Sender.MaxBackups)
	remoteHost := setupSSH(conf)
	cronChannel, cronjob := setupCron(conf.System.Role.Sender.Cron)
	ticker, testSSH := setupTicker(60)
	ex := setupExec()

	// database dump prep and manipulation
	//   - get the existing backups
	//   - parse the multiline string into an array
	//   - strip path
	//   - add sorted backups to ring buffer
	//   - delete backups that didn't fit into ring buffer
	//   - start ticker that monitors ssh connection

	if err := remoteHost.Connect(); err != nil {
		return err
	}
	remoteAlive := true
	backupsAsString, err := backup.GetBackups(remoteHost, conf.System.Role.Sender.DBname, conf.System.WorkingDir, ex)
	if err != nil {
		return err
	}
	backups, err := parseMultilineString(backupsAsString)
	if err != nil {
		glog.Fatal(err)
	}
	backups = StripPath(conf.System.WorkingDir, backups)
	expiredDumps := fillBuffer(buff, SortBackups(backups))
	if err := backup.Delete(remoteHost, conf.System.WorkingDir, expiredDumps, ex); err != nil {
		glog.Error(err)
	}
	cronjob.Start()
	startTicker(ticker, testSSH)

	for {
		select {

		// test ssh connection
		case <-testSSH:

			if err = remoteHost.TestConnection(); err != nil {
				glog.Error(err)
			} else {
				break
			}
			glog.Error("remote connection is down. backups are suspended until connection is re-established")
			if err := remoteHost.Reconnect(3, 10); err != nil {
				glog.Error(err)
				remoteAlive = false
			} else {
				remoteAlive = true
			}

		// cron trigger
		case <-cronChannel:
			if !remoteAlive {
				glog.Error("remote is down")
				break
			}

			dumpStdout, err := dB.Dump()
			if err != nil {
				glog.Error(err)
				break
			}

			err = backup.ToRemote(remoteHost, conf.System.WorkingDir, dB.Name(), dumpStdout, ex)
			if err != nil {
				glog.Error(err)
				break
			}
			expiredDump := buff.Enqueue(dB.Name())
			if expiredDump == "" {
				break
			}
			if err := backup.Delete(remoteHost, conf.System.WorkingDir, []string{expiredDump}, ex); err != nil {
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

func setupSSH(conf *conf.Config) *inet.SSH {

	// setup remote SSH connection
	var remoteConnPtr = new(inet.SSH)
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

func fillBuffer(buff *CircularQueue, sortedBackups []string) []string {

	expiredBuffElements := buff.Populate(sortedBackups)

	for _, elem := range buff.queue[0:buff.size] {
		glog.Info("existing backups: " + elem.name)
	}

	return expiredBuffElements
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

func setupExec() *exec.Exec {

	var ex = new(exec.Exec)
	return ex
}

// factory to setup chosen database
func setupSenderDB(impl string, ip net.IPAddr, port uint16, user string, pass string, name string) db.DB {
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
