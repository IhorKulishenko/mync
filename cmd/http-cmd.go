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
	url      string
	verb     string
	output   string
	postBody string
}

type argKeys struct {
	verb       string
	outputFile string
	body       string
	bodyFile   string
	wasSet     map[string]bool
}

func validate(keys argKeys) error {
	allowedMethods := []string{"GET", "POST", "HEAD"}
	if !slices.Contains(allowedMethods, strings.ToUpper(keys.verb)) {
		return ErrInvalidHttpMethod
	}

	if strings.ToUpper(keys.verb) != "POST" && (keys.wasSet["body"] || keys.wasSet["body-file"]) {
		return ErrInvalidHttpUsage
	}

	if keys.wasSet["body"] && keys.wasSet["body-file"] {
		return ErrInvalidHttpUsage
	}

	return nil
}

func HandleHttp(writer io.Writer, args []string) error {
	keys := argKeys{}

	fs := flag.NewFlagSet("http", flag.ContinueOnError)
	fs.SetOutput(writer)
	fs.StringVar(&keys.verb, "verb", "GET", "HTTP method")
	fs.StringVar(&keys.outputFile, "output", "", "output file name")
	fs.StringVar(&keys.body, "body", "", "POST body")
	fs.StringVar(&keys.bodyFile, "body-file", "", "POST body in file")

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

	keys.wasSet = make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		keys.wasSet[f.Name] = true
	})

	if err := validate(keys); err != nil {
		return err
	}

	c := httpConfig{
		verb:     strings.ToUpper(keys.verb),
		url:      fs.Arg(0),
		output:   keys.outputFile,
		postBody: getJsonBody(keys.body, keys.bodyFile),
	}

	return processVerb(writer, c)
}

func getJsonBody(fromString string, fromFile string) string {
	if len(fromFile) > 0 {
		fd, err := os.Open(fromFile)
		if err != nil {
			return ""
		}
		defer fd.Close()

		buffer, err := io.ReadAll(fd)
		if err != nil {
			return ""
		}

		return string(buffer)
	}

	return fromString
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
	case "POST":
		resp, err := postToRemoteSource(cfg.url, cfg.postBody)
		if err != nil {
			return err
		}
		data = resp

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

func postToRemoteSource(url string, json string) ([]byte, error) {

	resp, err := http.Post(url, "application/json", strings.NewReader(json))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
