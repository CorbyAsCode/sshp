package main

// From https://github.com/jilieryuyi/ssh-simple-client/blob/master/main.go
// link: http://blog.ralch.com/tutorial/golang-ssh-connection/

import (
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

const defaultTimeout = 20 // seconds

type sshClient struct {
	IP      string
	User    string
	Cert    string //password or key file path
	Port    int
	session *ssh.Session
	client  *ssh.Client
}

func (ssh_client *sshClient) readPublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func (ssh_client *sshClient) Connect() error {

	var sshConfig *ssh.ClientConfig
	var auth []ssh.AuthMethod
	auth = []ssh.AuthMethod{ssh_client.readPublicKeyFile(ssh_client.Cert)}

	sshConfig = &ssh.ClientConfig{
		User: ssh_client.User,
		Auth: auth,
		// HostKeyCallback
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: time.Second * defaultTimeout,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ssh_client.IP, ssh_client.Port), sshConfig)
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return err
	}

	ssh_client.session = session
	ssh_client.client = client
	return nil
}

func (ssh_client *sshClient) RunCmd(cmd string) ([]byte, error) {
	out, err := ssh_client.session.Output(cmd)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (ssh_client *sshClient) Close() {
	ssh_client.session.Close()
	ssh_client.client.Close()
}
