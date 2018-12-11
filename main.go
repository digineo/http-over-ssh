package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

var home = func() string {
	user, err := user.Current()
	if err == nil && user.HomeDir != "" {
		return user.HomeDir
	}
	return os.Getenv("HOME")
}()

var sshKeys = []string{
	filepath.Join(home, ".ssh/id_rsa"),
	filepath.Join(home, ".ssh/id_ed25519"),
}

// command line flags
var (
	listen     = "[::1]:8080"
	sshUser    = "root"
	sshTimeout = 10 * time.Second
)

func main() {
	flag.StringVar(&listen, "listen", listen, "listen on")
	flag.StringVar(&sshUser, "user", sshUser, "default SSH username")
	flag.DurationVar(&sshTimeout, "timeout", sshTimeout, "SSH connection timeout")
	flag.Parse()

	log.SetFlags(log.Lshortfile)

	authMethods := readPrivateKeys(sshKeys...)
	if len(authMethods) == 0 {
		log.Panicln("no SSH keys found")
	}

	proxy := NewProxy()
	proxy.sshConfig = ssh.ClientConfig{
		Timeout: sshTimeout,
		User:    sshUser,
		Auth:    authMethods,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// TODO implement this
			return nil
		},
	}

	log.Println("listening on", listen)
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		log.Panic(err)
	}

	log.Fatal(http.Serve(listener, proxy))
}
