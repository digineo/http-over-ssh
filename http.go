package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const defaultPort = 22

func (proxy *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target, err := url.Parse(r.RequestURI)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "error parsing URL:", err)
		return
	}

	if r.Host == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "hostname missing:", err)
		return
	}

	// parts[0] = ignored, should be empty
	// parts[1] = destination address
	parts := strings.SplitN(target.Path, "/", 2)
	if len(parts) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	key, err := parseJumpHost(target.Host)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "error parsing jump host:", err)
		return
	}
	if key.port == 0 {
		key.port = defaultPort
	}

	// get client
	client, err := proxy.getClient(key)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintln(w, err.Error())
		return
	}

	// build a new request
	req, err := http.NewRequest(r.Method, target.Scheme+"://"+parts[1], nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "unable to build request:", err)
		return
	}

	// set body
	req.Body = r.Body

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
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
