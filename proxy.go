package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"sync"

	"golang.org/x/crypto/ssh"
)

// Proxy holds the HTTP client and the SSH connection pool
type Proxy struct {
	httpClient *http.Client
	sshClients map[string]*ssh.Client
	sshConfig  *ssh.ClientConfig
	mtx        sync.Mutex
}

// NewProxy creates a new proxy
func NewProxy() *Proxy {
	proxy := &Proxy{
		sshClients: make(map[string]*ssh.Client),
	}
	proxy.httpClient = &http.Client{
		Transport: &http.Transport{
			DialContext: proxy.dialContext,
			DialTLS: func(network, addr string) (net.Conn, error) {
				return nil, errors.New("not implemented")
			},
		},
	}

	return proxy
}

// dialContext is used by the HTTP client
func (proxy *Proxy) dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	proxyCtx := ctx.(*proxyContext)

	sshClient, err := proxyCtx.getClient()
	if err != nil {
		return nil, err
	}

	// TODO: DialContext
	return sshClient.Dial(network, addr)
}

// getClient returns a connected SSH client
func (proxy *Proxy) getClient(address string) (*ssh.Client, error) {
	proxy.mtx.Lock()
	defer proxy.mtx.Unlock()

	// connection established?
	client, _ := proxy.sshClients[address]
	if client != nil {
		return client, nil
	}

	// try to connect
	log.Println("establishing SSH connection to", address)
	conn, err := ssh.Dial("tcp", address, proxy.sshConfig)
	if err != nil {
		return nil, err
	}

	// set and return the new connection
	proxy.sshClients[address] = conn
	return conn, nil
}
