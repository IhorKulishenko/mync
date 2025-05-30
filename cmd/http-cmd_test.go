package cmd

import (
	"bytes"
	"errors"
	"os"
	"testing"

	testutils "example.com/sub-cmd-example/test_utils"
)

func TestHandleHttp(t *testing.T) {
	usageMessage := `
http: A HTTP client.
http: <options> server

Options:
  -output string
    	output file name
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

func TestProcessVerb(t *testing.T) {
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
