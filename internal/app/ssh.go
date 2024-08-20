package app

import (
	"os"

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
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}
