package cmd

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"slices"
	"strings"
)

type httpConfig struct {
	disableRedirect bool
	url             string
	verb            string
	output          string
	postBody        string
	formData        map[string]string
	mpfile          string
}

type argKeys struct {
	body            string
	bodyFile        string
	disableRedirect bool
	formData        formDataArg
	outputFile      string
	uploadFile      string
	verb            string
	wasSet          map[string]bool
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
	keys := argKeys{formData: make(formDataArg)}

	fs := flag.NewFlagSet("http", flag.ContinueOnError)
	fs.SetOutput(writer)
	fs.StringVar(&keys.verb, "verb", "GET", "HTTP method")
	fs.StringVar(&keys.outputFile, "output", "", "output file name")
	fs.StringVar(&keys.body, "body", "", "POST body")
	fs.StringVar(&keys.bodyFile, "body-file", "", "POST body in file")
	fs.StringVar(&keys.uploadFile, "upload", "", "POST multipart form file upload")
	fs.Var(&keys.formData, "form-data", "POST multipart form data (key=value)")
	fs.BoolVar(&keys.disableRedirect, "disable-redirect", false, "GET disable redirect")
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

	cfg := httpConfig{
		verb:            strings.ToUpper(keys.verb),
		url:             fs.Arg(0),
		output:          keys.outputFile,
		postBody:        getJsonBody(keys.body, keys.bodyFile),
		formData:        keys.formData,
		mpfile:          keys.uploadFile,
		disableRedirect: keys.disableRedirect,
	}

	return processVerb(writer, cfg)
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

	client := getHttpClient(cfg)

	switch cfg.verb {
	case "GET":
		data, err = getRemoteResource(client, &cfg)
		if err != nil {
			return err
		}
	case "POST":
		var resp []byte
		var err error
		if len(cfg.formData) > 0 || len(cfg.mpfile) > 0 {
			resp, err = postMultiPartToRemoteSource(&cfg)
		} else {
			resp, err = postBodyToRemoteSource(&cfg)
		}

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

func getHttpClient(cfg httpConfig) *http.Client {
	if cfg.disableRedirect {
		return &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) > 0 {
					return errors.New("no redirects allowed")
				}

				return nil
			}}
	}

	return http.DefaultClient
}

func getRemoteResource(client *http.Client, cfg *httpConfig) ([]byte, error) {
	resp, err := client.Get(cfg.url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func postBodyToRemoteSource(cfg *httpConfig) ([]byte, error) {

	resp, err := http.Post(cfg.url, "application/json",
		strings.NewReader(cfg.postBody))

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func postMultiPartToRemoteSource(cfg *httpConfig) ([]byte, error) {
	var buffer = new(bytes.Buffer)

	mwriter := multipart.NewWriter(buffer)

	errResponse := func(err error) ([]byte, error) {
		return []byte{}, err
	}

	for k, v := range cfg.formData {
		fw, err := mwriter.CreateFormField(k)
		if err != nil {
			return errResponse(err)
		}
		fmt.Fprint(fw, v)
	}

	if len(cfg.mpfile) > 0 {
		fw, err := mwriter.CreateFormFile("file", cfg.mpfile)
		if err != nil {
			return errResponse(err)
		}

		freader, err := os.Open(cfg.mpfile)
		if err != nil {
			return errResponse(err)
		}

		defer freader.Close()
		_, err = io.Copy(fw, freader)
		if err != nil {
			return errResponse(err)
		}
	}

	err := mwriter.Close()
	if err != nil {
		return errResponse(err)
	}

	contentType := mwriter.FormDataContentType()

	resp, err := http.Post(cfg.url, contentType, buffer)
	if err != nil {
		return errResponse(err)
	}

	return io.ReadAll(resp.Body)
}
