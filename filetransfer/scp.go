package filetransfer

import (
	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
)

func Open (dest string, port string, username string, password string) scp.Client {

	config, _ := auth.PasswordKey(username, password, ssh.InsecureIgnoreHostKey())

	sc := scp.NewClient(dest + ":" + port, &config)

	err := sc.Connect()
	if err != nil {
		log.Fatal(err)
	}

	return sc
}

func Send (sc scp.Client, filename string, fileDir string) {

	fd, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	err = sc.CopyFile(fd, fileDir + filename, "0600")
	if err != nil {
		log.Fatal(err)
	}

	defer sc.Close()
	defer fd.Close()
}

func Cleanup (filename string) {

	err := os.Remove(filename)
	if err != nil {
		log.Fatal(err)
	}
}
