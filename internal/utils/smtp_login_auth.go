package utils

import (
	"fmt"
	"net/smtp"
	"strings"
)

type loginAuth struct {
	username string
	password string
	host     string
}

// LoginAuth creates AUTH LOGIN mechanism (Office365 compatible)
func LoginAuth(username, password, host string) smtp.Auth {
	return &loginAuth{
		username: username,
		password: password,
		host:     host,
	}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if server.Name != a.host {
		return "", nil, fmt.Errorf("wrong host name: %s", server.Name)
	}
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}

	challenge := strings.ToLower(string(fromServer))
	switch {
	case strings.Contains(challenge, "username"):
		return []byte(a.username), nil
	case strings.Contains(challenge, "password"):
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("unexpected server challenge: %s", fromServer)
	}
}
