package cmd

import (
	"bytes"
	"errors"
	"testing"

	testutils "example.com/sub-cmd-example/test_utils"
)

func TestHandleHttp(t *testing.T) {
	usageMessage := `
http: A HTTP client.
http: <options> server

Options:
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
