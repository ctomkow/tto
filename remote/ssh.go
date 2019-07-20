package remote

import (
	"golang.org/x/crypto/ssh"
	"log"
)

type SSH struct {
	remoteHostName string
	remoteHostPort string
	user string
	pass string
	config *ssh.ClientConfig
	session *ssh.Session
	connection ssh.Conn
}

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

func (sc *SSH) Connect() {
	client, err := ssh.Dial("tcp", sc.remoteHostName + ":" + sc.remoteHostPort, sc.config)
	if err != nil {
		log.Fatal(err)
	}

	sc.connection = client.Conn
	sc.session, err = client.NewSession()
	if err != nil {
		log.Fatal(err)
	}
}
