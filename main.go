package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

func main() {
	listen := flag.String("listen", ":8080", "listen on")
	sshUser := flag.String("user", "root", "default SSH username")
	sshTimeout := flag.Duration("timeout", time.Second*10, "SSH connection timeout")
	flag.Parse()

	home := os.Getenv("HOME")
	keys := []string{
		home + "/.ssh/id_rsa",
		home + "/.ssh/id_ed25519",
	}

	log.SetFlags(log.Lshortfile)

	authMethods := readPrivateKeys(keys...)
	if len(authMethods) == 0 {
		log.Panicln("no SSH keys found")
	}

	proxy := NewProxy()
	proxy.sshConfig = &ssh.ClientConfig{
		Timeout: *sshTimeout,
		User:    *sshUser,
		Auth:    authMethods,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// TODO implement this
			return nil
		},
	}

	log.Println("listening on", *listen)
	listener, err := net.Listen("tcp", *listen)
	if err != nil {
		panic(err)
	}

	http.Serve(listener, proxy)
}
