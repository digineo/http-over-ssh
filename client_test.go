package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRequest(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		requestURI    string
		authorization string
		expectedKey   clientKey
		expectedURI   string
		expectedError string // only the message
	}{
		// URI without host
		{
			requestURI:    "/",
			expectedError: "host missing in request URI",
		},
		// URI without destination host
		{
			requestURI:    "http://example.com/",
			expectedError: "destination host missing in request URI",
		},
		// URI without slash after target host
		{
			requestURI:  "http://example.com/localhost",
			expectedKey: clientKey{host: "example.com", port: 22},
			expectedURI: "http://localhost",
		},
		// Hostname without port
		{
			requestURI:    "http://example.com/localhost/metrics?foo=bar",
			authorization: "Basic cHJvbWV0aGV1czo=",
			expectedKey:   clientKey{host: "example.com", port: 22, username: "prometheus"},
			expectedURI:   "http://localhost/metrics?foo=bar",
		},
		// IPv4 address with port
		{
			requestURI:    "http://127.0.0.1:22/localhost:9100/",
			authorization: "Basic dXNlcjpzZWNyZXQ=",
			expectedKey:   clientKey{host: "127.0.0.1", port: 22, username: "user"},
			expectedURI:   "http://localhost:9100/",
		},
		// IPv6 with port
		{
			requestURI:  "http://[fe80::1]:2222/[fe80::2]:9100/metrics",
			expectedKey: clientKey{host: "fe80::1", port: 2222},
			expectedURI: "http://[fe80::2]:9100/metrics",
		},
	}

	for _, test := range tests {
		r := http.Request{
			RequestURI: test.requestURI,
			Header:     http.Header{},
		}

		if test.authorization != "" {
			r.Header.Add("Authorization", test.authorization)
		}

		key, uri, err := parseRequest(&r)

		if test.expectedError != "" {
			assert.EqualError(err, test.expectedError)
		} else {
			assert.NoError(err)
			assert.Equal(&test.expectedKey, key)
			assert.Equal(test.expectedURI, uri)
		}
	}
}
