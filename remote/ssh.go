// Craig Tomkow
// July 24, 2019

package remote

import (
	"golang.org/x/crypto/ssh"
)

// ##### structs #####

type SSH struct {
	remoteHostName string
	remoteHostPort string
	user string
	pass string
	config *ssh.ClientConfig
	session *ssh.Session
	connection *ssh.Client
}

// ##### public functions #####

// TODO: support for keys
func ConnPrep(ip string, port string, user string, pass string) *SSH {

	conf := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return &SSH{
		remoteHostName: ip,
		remoteHostPort: port,
		user:           user,
		pass:           pass,
		config:         conf,
	}
}

func (sc *SSH) Connect() error {
	client, err := ssh.Dial("tcp", sc.remoteHostName + ":" + sc.remoteHostPort, sc.config)
	if err != nil {
		return err
	}

	sc.connection = client
	sc.session, err = client.NewSession()
	if err != nil {
		return err
	}

	return nil
}

func (sc *SSH) NewSession() error {

	var err error
	sc.session, err = sc.connection.NewSession()
	if err != nil {
		return err
	}

	return nil
}

func (sc *SSH) CloseConnection() error {
	if err := sc.connection.Close(); err != nil {
		return err
	}

	return nil
}

func (sc *SSH) CloseSession() error {
	if err := sc.session.Close(); err != nil {
		return err
	}

	return nil
}
