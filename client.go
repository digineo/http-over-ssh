package main

import (
	"context"
	"errors"
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
	sshCert    *ssh.Certificate
	sshConfig  ssh.ClientConfig
	sshClient  *ssh.Client
	httpClient *http.Client
	mtx        sync.Mutex
}

// clientKey is used for reusing SSH connections.
type clientKey struct {
	host     string
	port     uint16
	username string
}

// hostPort returns the host joined with the port.
func (key *clientKey) hostPort() string {
	return net.JoinHostPort(key.host, strconv.Itoa(int(key.port)))
}

func (key *clientKey) String() string {
	hp := key.hostPort()
	if key.username == "" {
		return hp
	}
	return fmt.Sprintf("%s@%s", key.username, hp)
}

// establishes the SSH connection and sets up the HTTP client.
func (client *client) connect() error {
	sshClient, err := ssh.Dial("tcp", client.key.hostPort(), &client.sshConfig)
	if err != nil {
		metrics.connections.failed.Add(1)
		log.Printf("SSH connection to %s failed: %v", client.key.String(), err)
		return err
	}

	client.sshClient = sshClient
	metrics.connections.established.Add(1)
	log.Printf("SSH connection to %s established", client.key.String())

	return nil
}

// establishes a TCP connection through SSH.
func (client *client) dial(ctx context.Context, network, address string) (net.Conn, error) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	var conn net.Conn
	var err error

	for retried := false; ; retried = true {
		if client.sshClient == nil {
			if err := client.connect(); err != nil {
				return nil, err
			}
		}

		conn, err = client.sshClient.DialContext(ctx, network, address)
		if err == nil || retried || (!errors.Is(err, io.EOF) && client.isAlive()) {
			break
		}

		// ssh connection broken
		_ = client.sshClient.Close()
		client.sshClient = nil

		// Clean up idle HTTP connections
		client.httpClient.Transport.(*http.Transport).CloseIdleConnections()
	}

	if err == nil {
		metrics.forwardings.established.Add(1)
		log.Printf("TCP forwarding via %s to %s established", client.key.String(), address)
	} else {
		metrics.forwardings.failed.Add(1)
		log.Printf("TCP forwarding via %s to %s failed: %s", client.key.String(), address, err)
	}

	return conn, err
}

// checks if the SSH client is still alive by sending a keep alive request.
func (client *client) isAlive() bool {
	_, _, err := client.sshClient.SendRequest("keepalive@openssh.com", true, nil)

	return err == nil
}
