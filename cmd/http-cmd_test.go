package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	testutils "example.com/sub-cmd-example/test_utils"
)

func TestHandleHttp(t *testing.T) {
	usageMessage := `
http: A HTTP client.
http: <options> server

Options:
  -body string
    	POST body
  -body-file string
    	POST body in file
  -form-data value
    	POST multipart form data (key=value)
  -output string
    	output file name
  -upload string
    	POST multipart form file upload
  -verb string
    	HTTP method (default "GET")
`
	expectedResponse := "test response"
	ts := testutils.StartTestHttpServer(expectedResponse)
	defer ts.Close()

	testConfigs := []struct {
		args   []string
		output string
		err    error
	}{
		{
			args: []string{},
			err:  ErrNoServerSpecified,
		},
		{
			args:   []string{"-h"},
			err:    errors.New("flag: help requested"),
			output: usageMessage,
		},
		{
			args:   []string{ts.URL},
			err:    nil,
			output: expectedResponse,
		},
		{
			args:   []string{"-verb", "GET", ts.URL},
			err:    nil,
			output: expectedResponse,
		},
		{
			args: []string{"-verb", "PUT", ts.URL},
			err:  ErrInvalidHttpMethod,
		},

		{
			args: []string{"-verb", "GET", "http://nonexistent_url"},
			err:  errors.New("Get \"http://nonexistent_url\": dial tcp: lookup nonexistent_url: no such host"),
		},
	}

	byteBuf := new(bytes.Buffer)
	for index, tc := range testConfigs {
		err := HandleHttp(byteBuf, tc.args)

		if tc.err == nil && err != nil {
			t.Fatalf("T%d: Expected nil error, got %v", index, err)
		}

		if tc.err != nil && err != nil && tc.err.Error() != err.Error() {
			t.Fatalf("T%d: Expected error %v, got error: %v", index, tc.err, err)
		}

		if len(tc.output) != 0 {
			gotOutput := byteBuf.String()
			if tc.output != gotOutput {
				t.Errorf("T%d: Expected output %#v, got: %#v", index, tc.output, gotOutput)
			}
		}

		byteBuf.Reset()
	}
}

func TestProcessGetVerb(t *testing.T) {
	expectedResponse := "test response"
	ts := testutils.StartTestHttpServer(expectedResponse)
	defer ts.Close()

	testConfigs := []struct {
		name   string
		cfg    httpConfig
		err    error
		output string
	}{
		{
			name: "GET request success",
			cfg: httpConfig{
				verb: "GET",
				url:  ts.URL,
			},
			output: expectedResponse,
		},
		{
			name: "Write to output file",
			cfg: httpConfig{
				verb:   "GET",
				url:    ts.URL,
				output: t.TempDir() + "/test.out",
			},
			output: "",
		},
		{
			name: "Invalid output file",
			cfg: httpConfig{
				verb:   "GET",
				url:    ts.URL,
				output: "/invalid/path/test.out",
			},
			err: errors.New("open /invalid/path/test.out: no such file or directory"),
		},
	}

	buffer := new(bytes.Buffer)

	for _, tc := range testConfigs {
		err := processVerb(buffer, tc.cfg)

		if tc.err != nil {
			if err == nil || err.Error() != tc.err.Error() {
				t.Errorf("expected error %v, got %v", tc.err, err)
			}
			return
		}

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tc.output != "" && buffer.String() != tc.output {
			t.Errorf("expected output %q, got %q", tc.output, buffer.String())
		}

		if tc.cfg.output != "" {
			data, err := os.ReadFile(tc.cfg.output)
			if err != nil {
				t.Fatalf("error reading output file: %v", err)
			}
			if string(data) != expectedResponse {
				t.Errorf("file content mismatch, expected %q, got %q", expectedResponse, string(data))
			}
		}

		buffer.Reset()
	}
}

func TestProcessPostVerb(t *testing.T) {
	ts := testutils.StartExtendedTestHttpServer()
	defer ts.Close()

	testConfigs := []struct {
		name   string
		cfg    httpConfig
		err    error
		output string
	}{
		{
			name: "",
			cfg: httpConfig{
				url:      ts.URL,
				verb:     "POST",
				postBody: `{"value":"test value 1"}`,
			},
			output: "method: POST\nbody: \"{\\\"value\\\":\\\"test value 1\\\"}\"\n",
		},
	}

	bufferWriter := new(bytes.Buffer)
	for index, tc := range testConfigs {
		err := processVerb(bufferWriter, tc.cfg)

		if tc.err == nil && err != nil {
			t.Fatalf("T%d: Expected non error, got: %v\n", index, err)
		}

		if tc.err != nil && err == nil {
			t.Fatalf("T%d: Expected error %v, got no error\n", index, tc.err)
		}

		if tc.err != err {
			t.Fatalf("T%d: Expected error %v, got error %v\n", index, tc.err, err)
		}

		actualResponse := bufferWriter.String()
		if tc.output != actualResponse {
			t.Fatalf("T%d: Expected output %#v, got %#v\n", index, tc.output, actualResponse)
		}

		bufferWriter.Reset()
	}

	///
	jsonFile := filepath.Join(t.TempDir(), "file.json")

	expectedResponse := "method: POST\nbody: \"{\\\"value\\\": \\\"some value from file\\\"}\"\n"
	os.WriteFile(jsonFile, []byte("{\"value\": \"some value from file\"}"), 0644)

	HandleHttp(bufferWriter, []string{"-verb", "post", "-body-file", jsonFile, ts.URL})
	actualResponse := bufferWriter.String()

	if expectedResponse != actualResponse {
		t.Fatalf("Expected %#v, got: %#v\n", expectedResponse, actualResponse)
	}
}

func TestPostMultipartVerb(t *testing.T) {
	ts := testutils.StartExtendedTestHttpServer()
	defer ts.Close()

	jsonFile := filepath.Join(t.TempDir(), "file.json")
	os.WriteFile(jsonFile, []byte("{\"value\": \"some value from file\"}"), 0644)

	bufferWriter := new(bytes.Buffer)
	err := HandleHttp(bufferWriter, []string{"-verb", "post", "-form-data", "key1=value1", "-form-data", "key2=value2", "-upload", jsonFile, ts.URL})
	if err != nil {
		t.Fatalf("got error %v", err)
	}

	actualResponse := bufferWriter.String()
	expectedResponse := "value: key1=value1;key2=value2\nfiles: file=file.json;\n"

	if expectedResponse != actualResponse {
		t.Fatalf("Expected: %#v, got: %#v", expectedResponse, actualResponse)
	}

}
