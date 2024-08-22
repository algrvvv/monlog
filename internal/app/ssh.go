package app

import (
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

func NewSSHConfig(idRSA, user string) (*ssh.ClientConfig, error) {
	key, err := os.ReadFile(idRSA)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
			//метод подключения по паролю: ssh.Password("password here"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}, nil
}
