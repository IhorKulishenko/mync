package cmd

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
)

type httpConfig struct {
	url  string
	verb string
}

func validate(method string) error {
	allowedMethods := []string{"GET", "POST", "HEAD"}
	if slices.Contains(allowedMethods, strings.ToUpper(method)) {
		return nil
	}

	return ErrInvalidHttpMethod
}

func HandleHttp(writer io.Writer, args []string) error {
	var verb string

	fs := flag.NewFlagSet("http", flag.ContinueOnError)
	fs.SetOutput(writer)
	fs.StringVar(&verb, "verb", "GET", "HTTP method")
	fs.Usage = func() {
		var usageString = `
http: A HTTP client.
http: <options> server`

		fmt.Fprintln(writer, usageString)
		fmt.Fprintln(writer)
		fmt.Fprintln(writer, "Options:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return ErrNoServerSpecified
	}

	if err := validate(verb); err != nil {
		return err
	}

	c := httpConfig{verb: strings.ToUpper(verb)}
	c.url = fs.Arg(0)

	return processVerb(writer, c)
}

func processVerb(writer io.Writer, cfg httpConfig) error {
	switch cfg.verb {
	case "GET":
		var data []byte
		var err error
		if data, err = getRemoteResource(cfg.url); err != nil {
			return err
		}
		fmt.Fprint(writer, string(data))
	default:
		panic("not immplemented")
	}

	return nil
}

func getRemoteResource(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
