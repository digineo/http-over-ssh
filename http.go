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
	defer r.Body.Close()

	key, uri, err := parseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "invalid proxy request:", err)
		return
	}

	// get client
	client, err := proxy.getClient(*key)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintln(w, err.Error())
		return
	}

	// build a new request
	req, err := http.NewRequest(r.Method, uri, nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "unable to build request:", err)
		return
	}

	// set header and body
	req.Header = cloneHeader(r.Header)
	removeConnectionHeaders(req.Header)
	req.Body = r.Body

	// Remove hop-by-hop headers to the backend. Especially
	// important is "Connection" because we want a persistent
	// connection, regardless of what the client sent to us.
	for _, h := range hopHeaders {
		if req.Header.Get(h) != "" {
			req.Header.Del(h)
		}
	}

	// do the request
	res, err := client.httpClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintln(w, err.Error())
		return
	}

	// copy response header and body
	copyHeader(w.Header(), res.Header)
	w.WriteHeader(res.StatusCode)
	io.Copy(w, res.Body)
	res.Body.Close()
}

func parseRequest(r *http.Request) (*clientKey, string, error) {
	target, err := url.Parse(r.RequestURI)
	if err != nil {
		return nil, "", err
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
			return nil, "", fmt.Errorf("unable to parse port number: %v", err)
		}
		key.port = uint16(ui)
	} else {
		key.port = defaultPort
	}

	// Parse username
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Basic ") {
		decoded, _ := base64.StdEncoding.DecodeString(auth[6:])
		if i := bytes.IndexByte(decoded, ':'); i != -1 {
			key.username = string(decoded[:i])
		} else {
			key.username = string(decoded)
		}

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

func cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

// removeConnectionHeaders removes hop-by-hop headers listed in the "Connection" header of h.
// See RFC 2616, section 14.10.
func removeConnectionHeaders(h http.Header) {
	if c := h.Get("Connection"); c != "" {
		for _, f := range strings.Split(c, ",") {
			if f = strings.TrimSpace(f); f != "" {
				h.Del(f)
			}
		}
	}
}
