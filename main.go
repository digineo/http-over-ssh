package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/ssh"
)

var home = func() string {
	user, err := user.Current()
	if err == nil && user.HomeDir != "" {
		return user.HomeDir
	}
	return os.Getenv("HOME")
}()

var sshKeyDir = or(os.Getenv("HOS_KEY_DIR"), filepath.Join(home, ".ssh"))

var sshKeys = []string{
	filepath.Join(sshKeyDir, "id_rsa"),
	filepath.Join(sshKeyDir, "id_ed25519"),
}

// command line flags
var (
	listen     = or(os.Getenv("HOS_LISTEN"), "[::1]:8080")
	sshUser    = or(os.Getenv("HOS_USER"), "root")
	sshTimeout = func() time.Duration {
		dur := os.Getenv("HOS_TIMEOUT")
		if dur != "" {
			if d, err := time.ParseDuration(dur); err != nil {
				return d
			}
		}
		return 10 * time.Second
	}()
)

// build flags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	fmt.Printf("%s %v, commit %v, built at %v\n", os.Args[0], version, commit, date)

	flag.StringVar(&listen, "listen", listen, "listen on")
	flag.StringVar(&sshUser, "user", sshUser, "default SSH username")
	flag.DurationVar(&sshTimeout, "timeout", sshTimeout, "SSH connection timeout")
	flag.Parse()

	log.SetFlags(log.Lshortfile)

	authMethods := readPrivateKeys(sshKeys...)
	if len(authMethods) == 0 {
		log.Fatal("no SSH keys found")
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

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", proxy)

	log.Println("listening on", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}

func or(s, alt string) string {
	if s != "" {
		return s
	}
	return alt
}
