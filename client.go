package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"

	"golang.org/x/crypto/ssh"
)

type client struct {
	key        clientKey
	sshConfig  ssh.ClientConfig
	sshClient  *ssh.Client
	httpClient *http.Client
	mtx        sync.Mutex
}

type clientKey struct {
	address  string
	port     int
	username string
}

// hostPort returns the host joined with the port
func (key *clientKey) hostPort() string {
	return net.JoinHostPort(key.address, strconv.Itoa(key.port))
}

func (key *clientKey) String() string {
	hp := key.hostPort()
	if key.username == "" {
		return hp
	}
	return fmt.Sprintf("%s@%s", key.username, hp)
}

// establishes the SSH connection and sets up the HTTP client
func (client *client) connect() error {
	log.Printf("establishing SSH connection to %s", client.key.String())

	sshClient, err := ssh.Dial("tcp", client.key.hostPort(), &client.sshConfig)
	if err != nil {
		return err
	}

	client.sshClient = sshClient
	return nil
}

// establishes a TCP connection through SSH
func (client *client) dial(network, address string) (net.Conn, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	retried := false

retry:
	if client.sshClient == nil {
		if err := client.connect(); err != nil {
			return nil, err
		}
	}

	log.Printf("forwarding via %s to %s", client.key.String(), address)

	conn, err := client.sshClient.Dial(network, address)
	if err == io.EOF && !retried {
		// ssh connection broken
		client.sshClient.Close()
		client.sshClient = nil
		retried = true
		goto retry
	}

	return conn, err
}
