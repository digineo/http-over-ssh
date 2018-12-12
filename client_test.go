package main

import (
	"log"
	"testing"
)

func TestParseJumpHost(t *testing.T) {
	tt := map[string]struct {
		expected clientKey
		err      string // only the message
	}{
		// simple error cases
		"":         {err: "empty jump host"},
		":":        {err: `strconv.Atoi: parsing "": invalid syntax`},
		"[]":       {err: "empty jump host"},
		"[":        {err: "empty jump host"},
		"@":        {err: "empty jump host"},
		"user@:22": {err: "empty jump host"},
		"[fe80::1": {err: "address [fe80::1: missing ']' in address"},

		// TODO: should these be an error?
		"]":           {expected: clientKey{address: "]"}},
		"fe80::1]":    {expected: clientKey{address: "fe80::1]"}},
		"[xyz::1]:22": {expected: clientKey{address: "xyz::1", port: 22}},

		// host as domain name
		"example.com":         {expected: clientKey{address: "example.com"}},
		"example.com:22":      {expected: clientKey{address: "example.com", port: 22}},
		"user@example.com":    {expected: clientKey{address: "example.com", username: "user"}},
		"user@example.com:22": {expected: clientKey{address: "example.com", username: "user", port: 22}},

		// host as IPv4 address
		"127.0.0.1":         {expected: clientKey{address: "127.0.0.1"}},
		"127.0.0.1:22":      {expected: clientKey{address: "127.0.0.1", port: 22}},
		"user@127.0.0.1":    {expected: clientKey{address: "127.0.0.1", username: "user"}},
		"user@127.0.0.1:22": {expected: clientKey{address: "127.0.0.1", username: "user", port: 22}},

		// host as IPv6 address
		"::1":               {expected: clientKey{address: "::1"}},
		"[::1]":             {expected: clientKey{address: "::1"}},
		"[::1]:22":          {expected: clientKey{address: "::1", port: 22}},
		"fe80::1":           {expected: clientKey{address: "fe80::1"}},
		"fe80::1:22":        {expected: clientKey{address: "fe80::1:22"}}, // !
		"[fe80::1]:22":      {expected: clientKey{address: "fe80::1", port: 22}},
		"user@fe80::1":      {expected: clientKey{address: "fe80::1", username: "user"}},
		"user@[fe80::1]":    {expected: clientKey{address: "fe80::1", username: "user"}},
		"user@[fe80::1]:22": {expected: clientKey{address: "fe80::1", username: "user", port: 22}},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			if name == "[xyz::1]:22" {
				log.Println("start debugging")
			}

			actual, err := parseJumpHost(name)

			// is either empty, and the other not?
			if (tc.err == "") != (err == nil) {
				if tc.err == "" {
					t.Errorf("\n\texpected no error, got %v\n", err)
				} else {
					t.Errorf("\n\texpected error %s, got nil\n", tc.err)
				}
				return
			}

			// is expected error same as actual error?
			if tc.err != "" {
				if tc.err != err.Error() {
					t.Errorf("\n\texpected error %s, got %v\n", tc.err, err)
				}
				return
			}

			// compare parsed value
			if tc.expected.address != actual.address ||
				tc.expected.port != actual.port ||
				tc.expected.username != actual.username {
				t.Errorf("\n\texpected %#v\n\tgot %#v\n", tc.expected, actual)
			}
		})
	}
}
