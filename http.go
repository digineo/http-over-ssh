package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const defaultPort = 22

func (proxy *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	key, uri, err := parseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r.Close = false
	r.Host = ""
	r.URL, _ = url.Parse(uri)
	r.RequestURI = ""
	removeHopHeaders(r.Header)

	// do the request
	res, err := proxy.getClient(*key).httpClient.Do(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// copy response header and body
	copyHeader(w.Header(), res.Header)
	w.WriteHeader(res.StatusCode)
	_, _ = io.Copy(w, res.Body)
	_ = res.Body.Close()
}

func parseRequest(r *http.Request) (*clientKey, string, error) {
	target, err := url.Parse(r.RequestURI)
	if err != nil {
		return nil, "", fmt.Errorf("unable to parse URI: %w", err)
	}

	if target.Host == "" {
		return nil, "", errors.New("host missing in request URI")
	}

	if target.Path == "/" {
		return nil, "", errors.New("destination host missing in request URI")
	}

	key := clientKey{
		host: target.Hostname(),
	}

	// Parse port
	if port := target.Port(); port != "" {
		ui, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			return nil, "", fmt.Errorf("parsing \"%v\": invalid port number", port)
		}
		key.port = uint16(ui)
	} else {
		key.port = defaultPort
	}

	// Parse username
	if auth, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Basic "); ok {
		decoded, _ := base64.StdEncoding.DecodeString(auth)
		username, _, _ := bytes.Cut(decoded, []byte{':'})
		key.username = string(username)
	}

	return &key, target.Scheme + ":/" + target.RequestURI(), nil
}

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Proxy-Connection", // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // canonicalized version of "TE"
	"Trailer", // not Trailers per URL above; http://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	"Upgrade",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// removeHopHeaders removes hop-by-hop headers to the backend. Especially
// important is "Connection" because we want a persistent
// connection, regardless of what the client sent to us.
func removeHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}
