package cmd

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"
)

type httpConfig struct {
	disableRedirect bool
	url             string
	verb            string
	output          string
	postBody        string
	formData        map[string]string
	headers         map[string]string
	mpfile          string
}

type argKeys struct {
	body            string
	bodyFile        string
	disableRedirect bool
	formData        mKeyArg
	headers         mKeyArg
	outputFile      string
	url             string
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
	keys, err := parseKeys(writer, args)
	if err != nil {
		return err
	}

	if err := validate(keys); err != nil {
		return err
	}

	cfg := httpConfig{
		verb:            strings.ToUpper(keys.verb),
		url:             keys.url,
		output:          keys.outputFile,
		postBody:        getJsonBody(keys.body, keys.bodyFile),
		formData:        keys.formData,
		mpfile:          keys.uploadFile,
		disableRedirect: keys.disableRedirect,
		headers:         keys.headers,
	}

	client := getHttpClient(cfg)

	ctx, cancelFn := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelFn()

	request, err := getRequest(ctx, cfg)
	if err != nil {
		return err
	}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	err = processResponse(writer, cfg, resp)

	return err
}

func processResponse(writer io.Writer, cfg httpConfig, resp *http.Response) error {
	if len(cfg.output) > 0 {
		file, err := os.OpenFile(cfg.output, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer file.Close()

		writer = file
	}

	defer resp.Body.Close()

	_, err := io.Copy(writer, resp.Body)
	return err
}

func parseKeys(writer io.Writer, args []string) (argKeys, error) {
	keys := argKeys{
		formData: make(mKeyArg),
		headers:  make(mKeyArg),
	}

	fs := flag.NewFlagSet("http", flag.ContinueOnError)
	fs.SetOutput(writer)
	fs.StringVar(&keys.verb, "verb", "GET", "HTTP method")
	fs.StringVar(&keys.outputFile, "output", "", "output file name")
	fs.StringVar(&keys.body, "body", "", "POST body")
	fs.StringVar(&keys.bodyFile, "body-file", "", "POST body in file")
	fs.StringVar(&keys.uploadFile, "upload", "", "POST multipart form file upload")
	fs.Var(&keys.formData, "form-data", "POST multipart form data (key=value)")
	fs.BoolVar(&keys.disableRedirect, "disable-redirect", false, "GET disable redirect")
	fs.Var(&keys.headers, "header", "custom header (key=value)")

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
		return argKeys{}, err
	}

	if fs.NArg() != 1 {
		return argKeys{}, ErrNoServerSpecified
	}

	keys.wasSet = make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		keys.wasSet[f.Name] = true
	})

	keys.url = fs.Arg(0)

	return keys, nil
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

func getRequest(ctx context.Context, cfg httpConfig) (*http.Request, error) {
	contentType, content, err := getRequestBody(cfg)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, cfg.verb, cfg.url, content)
	req.Header.Add("Content-Type", contentType)

	for hName, hValue := range cfg.headers {
		req.Header.Add(hName, hValue)
	}

	return req, err
}

func getRequestBody(cfg httpConfig) (string, io.Reader, error) {
	if cfg.verb == http.MethodGet {
		return "plain/text", nil, nil
	}

	if cfg.verb == http.MethodPost {
		if len(cfg.formData) > 0 || len(cfg.mpfile) > 0 {
			contentType, buffer, err := getMultipartBody(&cfg)
			return contentType, buffer, err
		} else {
			return "application/json", strings.NewReader(cfg.postBody), nil
		}
	}

	return "", nil, errors.New("unsupported method")
}

func getMultipartBody(cfg *httpConfig) (string, io.Reader, error) {
	var buffer = new(bytes.Buffer)

	mwriter := multipart.NewWriter(buffer)

	errResponse := func(err error) (string, io.Reader, error) {
		return "", nil, err
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

	return contentType, buffer, nil
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
