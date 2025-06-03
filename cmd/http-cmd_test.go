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
  -disable-redirect
    	GET disable redirect
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
