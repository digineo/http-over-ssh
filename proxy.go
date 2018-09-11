package main

import (
	"errors"
	"log"
	"net"
	"net/http"
	"sync"

	"golang.org/x/crypto/ssh"
)

// Proxy holds the HTTP client and the SSH connection pool
type Proxy struct {
	clients   map[clientKey]*client
	sshConfig *ssh.ClientConfig
	mtx       sync.Mutex
}

type client struct {
	sshClient  *ssh.Client
	httpClient *http.Client
}

type clientKey struct {
	address  string
	username string
}

// NewProxy creates a new proxy
func NewProxy() *Proxy {
	return &Proxy{
		clients: make(map[clientKey]*client),
	}
}

// getClient returns a connected SSH client
func (proxy *Proxy) getClient(key clientKey) (*client, error) {
	proxy.mtx.Lock()
	defer proxy.mtx.Unlock()

	// connection established?
	pClient, _ := proxy.clients[key]
	if pClient != nil {
		return pClient, nil
	}

	// try to connect
	log.Printf("establishing SSH connection to %+v", key)

	// copy sshConfig and set username
	sshConfig := *proxy.sshConfig
	sshConfig.User = key.username

	conn, err := ssh.Dial("tcp", key.address, &sshConfig)
	if err != nil {
		return nil, err
	}

	pClient = &client{
		sshClient: conn,
		httpClient: &http.Client{
			Transport: &http.Transport{
				Dial: conn.Dial,
				DialTLS: func(network, addr string) (net.Conn, error) {
					return nil, errors.New("not implemented")
				},
			},
		},
	}

	// set and return the new connection
	proxy.clients[key] = pClient
	return pClient, nil
}
