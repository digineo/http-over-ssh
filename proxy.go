package main

import (
	"errors"
	"net"
	"net/http"
	"sync"

	"golang.org/x/crypto/ssh"
)

// Proxy holds the HTTP client and the SSH connection pool.
type Proxy struct {
	clients   map[clientKey]*client
	sshConfig ssh.ClientConfig
	mtx       sync.Mutex
}

// NewProxy creates a new proxy.
func NewProxy() *Proxy {
	return &Proxy{
		clients: make(map[clientKey]*client),
	}
}

// getClient returns a (un)connected SSH client.
func (proxy *Proxy) getClient(key clientKey) *client {
	proxy.mtx.Lock()
	defer proxy.mtx.Unlock()

	// connection established?
	pClient := proxy.clients[key]
	if pClient != nil {
		return pClient
	}

	pClient = &client{
		key:       key,
		sshConfig: proxy.sshConfig, // make copy
	}
	pClient.sshConfig.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		if err := proxy.sshConfig.HostKeyCallback(hostname, remote, key); err != nil {
			return err
		}
		if cert, ok := key.(*ssh.Certificate); ok && cert != nil {
			pClient.sshCert = cert
		}
		return nil
	}

	if key.username != "" {
		pClient.sshConfig.User = key.username
	}
	pClient.httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: pClient.dial,
			DialTLS: func(network, addr string) (net.Conn, error) {
				return nil, errors.New("not implemented")
			},
		},
	}

	// set and return the new connection
	proxy.clients[key] = pClient
	return pClient
}
