package cmd

import (
	"flag"
	"fmt"
	"io"
)

type httpConfig struct {
	url  string
	verb string
}

func validate(method string) error {
	allowedMethods := []string{"GET", "POST", "HEAD"}
	for _, a := range allowedMethods {
		if method == a {
			return nil
		}
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

	c := httpConfig{verb: verb}
	c.url = fs.Arg(0)
	fmt.Fprintln(writer, "Executing http command")

	return nil
}
