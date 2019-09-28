// Craig Tomkow
// July 24, 2019

package inet

import (
	"errors"
	"github.com/golang/glog"
	"golang.org/x/crypto/ssh"
	"strconv"
	"time"
)

type SSH struct {
	remoteHostName string
	remoteHostPort string
	user           string
	pass           string
	config         *ssh.ClientConfig
	Session        *ssh.Session
	connection     *ssh.Client
}

// TODO: support for keys
func (sh *SSH) Make(ip string, port string, user string, pass string) {

	sh.remoteHostName = ip
	sh.remoteHostPort = port
	sh.user = user
	sh.pass = pass

	sh.config = &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
}

func (sh *SSH) Connect() error {
	client, err := ssh.Dial("tcp", sh.remoteHostName+":"+sh.remoteHostPort, sh.config)
	if err != nil {
		return err
	}

	sh.connection = client

	return nil
}

func (sh *SSH) NewSession() error {

	var err error
	sh.Session, err = sh.connection.NewSession()
	if err != nil {
		return err
	}

	return nil
}

func (sh *SSH) CloseConnection() error {
	if err := sh.connection.Close(); err != nil {
		return err
	}

	return nil
}

func (sh *SSH) CloseSession() error {
	if err := sh.Session.Close(); err != nil {
		return err
	}

	return nil
}

func (sh *SSH) GetSession() *ssh.Session {

	return sh.Session
}

func (sh *SSH) TestConnection() error {

	if err := sh.NewSession(); err != nil {
		return err
	}
	if err := sh.CloseSession(); err != nil {
		return errors.New("could not close test ssh Session")
	}

	return nil
}

func (sh *SSH) Reconnect(tries int, delayInSec int) error {

	for i := 1; i <= tries; i++ {
		glog.Error("[" + strconv.Itoa(i) + "/" + strconv.Itoa(tries) + "]" + " attempting to re-connect with remote")
		if err := sh.Connect(); err != nil {
			glog.Error("failed to re-establish connection with remote")
		} else {
			glog.Info("re-established connection with remote")
			return nil
		}

		time.Sleep(time.Duration(delayInSec) * time.Second)
	}

	return errors.New("reconnection with remote failed")
}
