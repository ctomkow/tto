// Craig Tomkow
// July 24, 2019

package remote

import (
	"golang.org/x/crypto/ssh"
)

type SSH struct {
	remoteHostName string
	remoteHostPort string
	user           string
	pass           string
	config         *ssh.ClientConfig
	session        *ssh.Session
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
	sh.session, err = sh.connection.NewSession()
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
	if err := sh.session.Close(); err != nil {
		return err
	}

	return nil
}
