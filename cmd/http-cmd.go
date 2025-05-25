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

func HandleHttp(writer io.Writer, args []string) error {
	var v string

	fs := flag.NewFlagSet("http", flag.ContinueOnError)
	fs.SetOutput(writer)
	fs.StringVar(&v, "verb", "GET", "HTTP method")
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

	c := httpConfig{verb: v}
	c.url = fs.Arg(0)
	fmt.Fprintln(writer, "Executing http command")

	return nil
}
