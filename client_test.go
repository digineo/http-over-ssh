package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		requestURI    string
		authorization string
		expectedKey   clientKey
		expectedURI   string
		expectedError string // only the message
	}{
		{
			name:          "URI without host",
			requestURI:    "/",
			expectedError: "host missing in request URI",
		},
		{
			name:          "URI without destination host",
			requestURI:    "http://example.com/",
			expectedError: "destination host missing in request URI",
		},
		{
			name:          "Invalid port number",
			requestURI:    "http://example.com:99999/localhost",
			expectedError: "parsing \"99999\": invalid port number",
		},
		{
			name:          "URI without slash after target host",
			requestURI:    "http://example.com/localhost",
			authorization: "Basic dGVzdA==",
			expectedKey:   clientKey{host: "example.com", port: 22, username: "test"},
			expectedURI:   "http://localhost",
		},
		{
			name:          "Hostname without port and credentials",
			requestURI:    "http://example.com/localhost/metrics?foo=bar",
			authorization: "Basic cHJvbWV0aGV1czo=",
			expectedKey:   clientKey{host: "example.com", port: 22, username: "prometheus"},
			expectedURI:   "http://localhost/metrics?foo=bar",
		},
		{
			name:          "IPv4 address with port",
			requestURI:    "http://127.0.0.1:22/localhost:9100/",
			authorization: "Basic dXNlcjpzZWNyZXQ=",
			expectedKey:   clientKey{host: "127.0.0.1", port: 22, username: "user"},
			expectedURI:   "http://localhost:9100/",
		},
		{
			name:        "IPv6 with port",
			requestURI:  "http://[fe80::1]:2222/[fe80::2]:9100/metrics",
			expectedKey: clientKey{host: "fe80::1", port: 2222},
			expectedURI: "http://[fe80::2]:9100/metrics",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)

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
		})
	}
}

func TestClientKeyToString(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	tests := []struct {
		name     string
		input    clientKey
		expected string
	}{
		{
			name:     "host with port",
			input:    clientKey{host: "example.com", port: 22},
			expected: "example.com:22",
		},
		{
			name:     "host with port and username",
			input:    clientKey{host: "example.com", port: 22, username: "prometheus"},
			expected: "prometheus@example.com:22",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(test.expected, test.input.String())
		})
	}
}

func TestGetClient(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	proxy := NewProxy()
	proxy.sshConfig.User = "default"

	// default username
	{
		client := proxy.getClient(clientKey{host: "::1"})
		assert.Equal("default", client.sshConfig.User)
	}

	// override username
	{
		client := proxy.getClient(clientKey{host: "::1", username: "prometheus"})
		assert.Equal("prometheus", client.sshConfig.User)
	}
}

func TestClientDialHTTPS(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	proxy := NewProxy()
	client := proxy.getClient(clientKey{host: "::1"})

	response, err := client.httpClient.Get("https://example.com/")
	assert.EqualError(err, `Get "https://example.com/": not implemented`)

	if response != nil {
		response.Body.Close()
	}
}

func TestInvalidRequestURI(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	w := httptest.NewRecorder()
	r := &http.Request{
		RequestURI: "%zz",
		Body:       ioutil.NopCloser(bytes.NewReader(nil)),
	}

	NewProxy().ServeHTTP(w, r)

	res := w.Result()
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	assert.NoError(err)
	assert.Equal(400, res.StatusCode)
	assert.Equal(`unable to parse URI: parse "%zz": invalid URL escape "%zz"`+"\n", string(body))
}
