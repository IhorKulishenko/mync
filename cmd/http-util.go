package cmd

import (
	"fmt"
	"strings"
)

type basicAuth struct {
	user     string
	password string
}

func (ba basicAuth) String() string {
	return fmt.Sprintf("%s:%s", ba.user, ba.password)
}

func (ba *basicAuth) Set(s string) error {
	if len(s) == 0 || !strings.Contains(s, ":") {
		return ErrInvalidHttpUsage
	}

	tokens := strings.Split(s, ":")
	if len(tokens) != 2 {
		return ErrInvalidHttpUsage
	}

	ba.user = tokens[0]
	ba.password = tokens[1]

	return nil
}

type mKeyArg map[string]string

func (fd mKeyArg) String() string {
	return fmt.Sprintf("%s", map[string]string(fd))
}

func (fd mKeyArg) Len() int {
	return len(fd)
}

func (fd mKeyArg) Set(s string) error {
	if len(s) == 0 || !strings.Contains(s, "=") {
		return ErrInvalidHttpUsage
	}

	tokens := strings.Split(s, "=")
	if len(tokens) != 2 {
		return ErrInvalidHttpUsage
	}

	key := tokens[0]
	value := tokens[1]

	fd[key] = value

	return nil
}
