package main

import (
	"context"

	"golang.org/x/crypto/ssh"
)

type proxyContext struct {
	context.Context        // super context
	proxy           *Proxy // reference to the proxy
	jumpHost        string // address of the jump host
}

func (ctx *proxyContext) getClient() (*ssh.Client, error) {
	return ctx.proxy.getClient(ctx.jumpHost)
}
