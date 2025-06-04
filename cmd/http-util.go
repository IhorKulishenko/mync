package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
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

type LoggingClient struct {
	log *log.Logger
}

func (c LoggingClient) RoundTrip(r *http.Request) (*http.Response, error) {

	tBegin := time.Now()

	resp, err := http.DefaultTransport.RoundTrip(r)

	duration := time.Since(tBegin)

	c.log.Printf("request took %v\n", duration)

	return resp, err
}

func NewLogMiddleware(w io.Writer) http.RoundTripper {
	transposrt := LoggingClient{}
	transposrt.log = log.New(w, "", log.LstdFlags)

	return transposrt
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) > 0 {
		return errors.New("no redirects allowed")
	}

	return nil
}
