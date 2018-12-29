package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
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

func parseJumpHost(jmp string) (ck clientKey, err error) {
	if jmp == "" {
		err = errors.New("empty jump host")
		return
	}

	if pos := strings.IndexRune(jmp, '@'); pos >= 0 {
		if len(jmp) <= pos+1 {
			err = errors.New("empty jump host")
			return
		}
		ck.username = jmp[0:pos]
		jmp = jmp[pos+1:]
	}

	if l := len(jmp); jmp[0] == '[' {
		if l <= 2 {
			err = errors.New("empty jump host")
			return
		}
		if jmp[l-1] == ']' {
			// IPv6 address without port
			ck.address = jmp[1 : l-1]
			return
		}

		// with port
		var p string
		ck.address, p, err = net.SplitHostPort(jmp)
		if err != nil {
			return
		}
		ck.port, err = strconv.Atoi(p)
		if err != nil {
			return
		}
		return
	}

	col := strings.IndexRune(jmp, ':')
	if col < 0 || col != strings.LastIndexByte(jmp, byte(':')) {
		// IPv6 address?
		ck.address = jmp
		return
	}

	var p string
	ck.address, p, err = net.SplitHostPort(jmp)
	if err != nil {
		return
	}
	ck.port, err = strconv.Atoi(p)
	if err != nil {
		return
	}

	if ck.address == "" {
		err = errors.New("empty jump host")
	}
	return
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
		metrics.connections.failed++
		return err
	}

	client.sshClient = sshClient
	metrics.connections.established++
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

	if err == nil {
		metrics.forwardings.established++
	} else {
		metrics.forwardings.failed++
	}

	return conn, err
}
