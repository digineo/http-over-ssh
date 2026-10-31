package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

type directTCPPayload struct {
	Addr       string // To connect to
	Port       uint32
	OriginAddr string
	OriginPort uint32
}

var (
	httpServer  *http.Server
	lastRequest *http.Request // the last received request
)

func TestHTTP(t *testing.T) {
	log.SetFlags(log.Llongfile)

	require := require.New(t)
	assert := assert.New(t)

	assert.EqualValues(0, metrics.connections.established.Load())
	assert.EqualValues(0, metrics.connections.failed.Load())
	assert.EqualValues(0, metrics.forwardings.established.Load())
	assert.EqualValues(0, metrics.forwardings.failed.Load())

	prometheus.MustRegister(&metrics)
	defer prometheus.Unregister(&metrics)

	config := &ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			// accept anything
			return nil, nil
		},
	}

	private, err := getKeyFile("fixtures/id_ed25519")
	require.NoError(err)
	config.AddHostKey(private)

	sshListener := listenLocal(t)
	httpListener := listenLocal(t)
	proxyListener := listenLocal(t)

	sshPort := sshListener.Addr().String()
	httpPort := httpListener.Addr().String()
	proxyPort := proxyListener.Addr().String()

	authMethods := readPrivateKeys("fixtures/id_ed25519")

	proxy = NewProxy()
	proxy.sshConfig = ssh.ClientConfig{
		Timeout: time.Second,
		User:    "prometheus",
		Auth:    authMethods,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// always successful
			return nil
		},
	}

	go serveSSH(sshListener, config)
	go serveHTTP(httpListener)
	go serveProxy(proxyListener, proxy)

	proxyURL, _ := url.Parse(fmt.Sprintf("http://%s/", proxyPort))
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}

	// valid request via valid jumphost
	response, err := client.Get(fmt.Sprintf("http://%s/%s/test", sshPort, httpPort))
	assert.NoError(err)
	if response != nil {
		assert.Equal(httpPort, lastRequest.Host)
		assert.EqualValues(1, metrics.connections.established.Load())
		assert.EqualValues(1, metrics.forwardings.established.Load())
		assert.EqualValues(0, metrics.forwardings.failed.Load())
		assert.Equal(200, response.StatusCode)

		bytes, _ := io.ReadAll(response.Body)
		assert.Equal("Hello World", string(bytes))
		response.Body.Close()
	}

	// valid HTTPS request via valid jumphost
	{
		response, err := client.Get(fmt.Sprintf("https://%s/%s/test", sshPort, httpPort))
		assert.Error(err)
		if response != nil {
			response.Body.Close()
		}
	}

	// invalid request via valid jumphost
	{
		response, err := client.Get(fmt.Sprintf("http://%s/%s/test", sshPort, "localhost:10000"))
		assert.NoError(err)
		if response != nil {
			assert.EqualValues(1, metrics.connections.established.Load())
			assert.EqualValues(0, metrics.connections.failed.Load())
			assert.EqualValues(1, metrics.forwardings.established.Load())
			assert.EqualValues(1, metrics.forwardings.failed.Load())
			assert.Equal(http.StatusBadGateway, response.StatusCode)
			response.Body.Close()
		}
	}

	// request via invalid jumphost
	{
		response, err := client.Get(fmt.Sprintf("http://%s/%s/test", "localhost:23", "localhost:10000"))
		assert.NoError(err)
		if response != nil {
			assert.Equal(http.StatusBadGateway, response.StatusCode)
			assert.EqualValues(1, metrics.connections.established.Load())
			assert.EqualValues(1, metrics.connections.failed.Load())
			assert.EqualValues(1, metrics.forwardings.established.Load())
			assert.EqualValues(1, metrics.forwardings.failed.Load())
			response.Body.Close()
		}
	}

	// metrics request
	{
		response, err := http.Get(fmt.Sprintf("http://%s/metrics", proxyPort))
		assert.NoError(err)
		if response != nil {
			assert.Equal(200, response.StatusCode)
			bytes, _ := io.ReadAll(response.Body)
			assert.Contains(string(bytes), `sshproxy_connections_total{state="established"} 1`)
			response.Body.Close()
		}
	}

	// Close HTTP server and SSH connections, SSH connection should be re-established,
	// forwarding should fail
	{
		// Closing the listener does not close active connections
		httpServer.Close()
		for _, client := range proxy.clients {
			if client := client.sshClient; client != nil {
				client.Close()
			}
		}

		response, err := client.Get(fmt.Sprintf("http://%s/%s/test", sshPort, httpPort))
		assert.NoError(err)
		assert.Equal(http.StatusBadGateway, response.StatusCode)
		assert.EqualValues(2, metrics.connections.established.Load())
		assert.EqualValues(1, metrics.connections.failed.Load())
		assert.EqualValues(1, metrics.forwardings.established.Load())
		assert.EqualValues(2, metrics.forwardings.failed.Load())
		response.Body.Close()
	}
}

// listenLocal opens a listener on a random local port.
func listenLocal(t *testing.T) net.Listener {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	return listener
}

func serveHTTP(listener net.Listener) {
	mux := http.ServeMux{}
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		lastRequest = r
		fmt.Fprint(w, "Hello World")
	})

	httpServer = &http.Server{
		Handler: &mux,
	}
	httpServer.Serve(listener)
}

func serveProxy(listener net.Listener, proxy *Proxy) {
	mux := http.ServeMux{}
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", proxy)
	server := http.Server{
		Handler: &mux,
	}
	server.Serve(listener)
}

func serveSSH(listener net.Listener, config *ssh.ServerConfig) {
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection (%s)", err)
			return
		}

		// Before use, a handshake must be performed on the incoming net.Conn.
		sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, config)
		if err != nil {
			log.Printf("Failed to handshake (%s)", err)
			return
		}

		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		// Discard all global out-of-band Requests
		go ssh.DiscardRequests(reqs)
		// Accept all channels
		go handleChannels(chans)
	}
}

func handleChannels(chans <-chan ssh.NewChannel) {
	// Service the incoming Channel channel in go routine
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "direct-tcpip" {
		panic(fmt.Sprintf("unknown channel type: %s", t))
	}

	var payload directTCPPayload
	if err := ssh.Unmarshal(newChannel.ExtraData(), &payload); err != nil {
		panic(err)
	}

	ipaddr, err := net.ResolveIPAddr("ip", payload.Addr)
	if err != nil {
		log.Println("Could not resolve address:", err)
		newChannel.Reject(ssh.Prohibited, err.Error())
		return
	}

	rconn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   ipaddr.IP,
		Port: int(payload.Port),
		Zone: ipaddr.Zone,
	})
	if err != nil {
		log.Println("Could not dial remote:", err)
		newChannel.Reject(ssh.Prohibited, err.Error())
		return
	}

	connection, requests, err := newChannel.Accept()
	if err != nil {
		panic(err)
	}
	go ssh.DiscardRequests(requests)

	serve(connection, rconn)
}

func serve(cssh ssh.Channel, conn net.Conn) {
	closeAll := func() {
		cssh.Close()
		conn.Close()
	}

	var once sync.Once
	go func() {
		io.Copy(cssh, conn)
		once.Do(closeAll)
	}()
	go func() {
		io.Copy(conn, cssh)
		once.Do(closeAll)
	}()
}
