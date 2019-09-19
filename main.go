package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var home = func() string {
	user, err := user.Current()
	if err == nil && user.HomeDir != "" {
		return user.HomeDir
	}
	return os.Getenv("HOME")
}()

var (
	sshKeyDir = envStr("HOS_KEY_DIR", filepath.Join(home, ".ssh"))
	sshKeys   = []string{
		filepath.Join(sshKeyDir, "id_rsa"),
		filepath.Join(sshKeyDir, "id_ed25519"),
	}
	knownHosts = filepath.Join(sshKeyDir, "known_hosts")
	proxy      *Proxy
)

// command line flags
var (
	listen        = envStr("HOS_LISTEN", "[::1]:8080")
	enableMetrics = envStr("HOS_METRICS", "1") != "0"
	sshUser       = envStr("HOS_USER", "root")
	sshTimeout    = envDur("HOS_TIMEOUT", 10*time.Second)
)

// build flags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	fmt.Printf("%s %v, commit %v, built at %v\n", os.Args[0], version, commit, date)

	flag.BoolVar(&enableMetrics, "metrics", enableMetrics, "enable metrics")
	flag.StringVar(&listen, "listen", listen, "listen on")
	flag.StringVar(&sshUser, "user", sshUser, "default SSH username")
	flag.DurationVar(&sshTimeout, "timeout", sshTimeout, "SSH connection timeout")
	flag.Parse()

	log.SetFlags(log.Lshortfile)

	authMethods := readPrivateKeys(sshKeys...)
	if len(authMethods) == 0 {
		log.Fatal("no SSH keys found")
	}

	hostKeyCallback, err := knownhosts.New(knownHosts)
	if err != nil {
		log.Fatal(err)
	}

	proxy = NewProxy()
	proxy.sshConfig = ssh.ClientConfig{
		Timeout:         sshTimeout,
		User:            sshUser,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
	}

	if enableMetrics {
		prometheus.MustRegister(&metrics)
		http.Handle("/metrics", promhttp.Handler())
	}

	http.Handle("/", proxy)
	log.Println("listening on", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}

func envStr(name, fallback string) string {
	if s := os.Getenv(name); s != "" {
		return s
	}
	return fallback
}

func envDur(name string, fallback time.Duration) time.Duration {
	if dur, err := time.ParseDuration(os.Getenv(name)); err == nil {
		return dur
	}
	return fallback
}
