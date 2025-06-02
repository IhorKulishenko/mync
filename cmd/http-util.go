package cmd

import (
	"fmt"
	"strings"
)

type formDataArg map[string]string

func (fd formDataArg) String() string {
	return fmt.Sprintf("%s", map[string]string(fd))
}

func (fd formDataArg) Len() int {
	return len(fd)
}

func (fd formDataArg) Set(s string) error {
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
