package cmd

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
)

type httpConfig struct {
	url    string
	verb   string
	output string
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
	var outputFile string

	fs := flag.NewFlagSet("http", flag.ContinueOnError)
	fs.SetOutput(writer)
	fs.StringVar(&verb, "verb", "GET", "HTTP method")
	fs.StringVar(&outputFile, "output", "", "output file name")
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

	c := httpConfig{
		verb:   strings.ToUpper(verb),
		url:    fs.Arg(0),
		output: outputFile,
	}

	return processVerb(writer, c)
}

func processVerb(writer io.Writer, cfg httpConfig) error {
	var data []byte
	var err error

	switch cfg.verb {
	case "GET":
		data, err = getRemoteResource(cfg.url)
		if err != nil {
			return err
		}

	default:
		panic("not immplemented")
	}

	if len(cfg.output) > 0 {
		file, err := os.OpenFile(cfg.output, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer file.Close()

		writer = file
	}

	fmt.Fprint(writer, string(data))

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
