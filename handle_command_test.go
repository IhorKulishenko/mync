package main

import (
	"bytes"
	"testing"

	"example.com/sub-cmd-example/cmd"
	testutils "example.com/sub-cmd-example/test_utils"
)

func TestHandleCommand(t *testing.T) {

	usageMessage := `Usage: mync [http|grpc] -h

http: A HTTP client.
http: <options> server

Options:
  -body string
    	POST body
  -body-file string
    	POST body in file
  -output string
    	output file name
  -verb string
    	HTTP method (default "GET")

grpc: A gRPC client.
grpc: <options> server

Options:
  -body string
    	Body of request
  -method string
    	Method to call
`

	expectedResponse := "testing"
	ts := testutils.StartTestHttpServer(expectedResponse)
	defer ts.Close()

	testConfigs := []struct {
		args   []string
		output string
		err    error
	}{
		{
			args:   []string{},
			err:    errInvalidSubCommand,
			output: "invalid sub-command specified\n" + usageMessage,
		},
		{
			args:   []string{"-h"},
			err:    nil,
			output: usageMessage,
		},
		{
			args:   []string{"foo"},
			err:    errInvalidSubCommand,
			output: "invalid sub-command specified\n" + usageMessage,
		},
		{
			args:   []string{"-h"},
			err:    nil,
			output: usageMessage,
		},
		{
			args:   []string{"-help"},
			err:    nil,
			output: usageMessage,
		},
		{
			args: []string{"http"},
			err:  cmd.ErrNoServerSpecified,
		},
		{
			args:   []string{"http", ts.URL},
			err:    nil,
			output: expectedResponse,
		},
		{
			args: []string{"grpc"},
			err:  cmd.ErrNoServerSpecified,
		},
		{
			args:   []string{"grpc", "remote.server"},
			err:    nil,
			output: "Executing grpc command\n",
		},
	}

	byteBuf := new(bytes.Buffer)
	for index, tc := range testConfigs {
		err := handleCommand(byteBuf, tc.args)

		if tc.err == nil && err != nil {
			t.Fatalf("T%d: Expected nil error, got: %v", index, err)
		}

		if tc.err != nil && tc.err.Error() != err.Error() {
			t.Fatalf("T%d: Expected error %v, got: %v", index, tc.err, err)
		}

		if len(tc.output) != 0 {
			gotOutput := byteBuf.String()
			if tc.output != gotOutput {
				t.Errorf("T%d: Expected output to be: %#v, got: %#v", index, tc.output, gotOutput)
			}
		}

		byteBuf.Reset()
	}

}
