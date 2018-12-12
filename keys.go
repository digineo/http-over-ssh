package main

import (
	"io/ioutil"
	"log"

	"golang.org/x/crypto/ssh"
)

// Reads a SSH private key file
func getKeyFile(path string) (ssh.Signer, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(buf)
}

func readPrivateKeys(paths ...string) (methods []ssh.AuthMethod) {
	for _, path := range paths {
		if signer, err := getKeyFile(path); err == nil {
			log.Println("loaded private key", path)
			methods = append(methods, ssh.PublicKeys(signer))
		} else {
			log.Printf("unable to load private key %q: %v", path, err)
		}
	}
	return methods
}
