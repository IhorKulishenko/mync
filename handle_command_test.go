package main

import (
	"bytes"
	"testing"

	"example.com/sub-cmd-example/cmd"
)

func TestHandleCommand(t *testing.T) {

	usageMessage := `Usage: mync [http|grpc] -h

http: A HTTP client.
http: <options> server

Options:
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
			args:   []string{"http", "remote.server"},
			err:    nil,
			output: "Executing http command\n",
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
	for _, tc := range testConfigs {
		err := handleCommand(byteBuf, tc.args)

		if tc.err == nil && err != nil {
			t.Fatalf("Expected nil error, got: %v", err)
		}

		if tc.err != nil && tc.err.Error() != err.Error() {
			t.Fatalf("Expected error %v, got: %v", tc.err, err)
		}

		if len(tc.output) != 0 {
			gotOutput := byteBuf.String()
			if tc.output != gotOutput {
				t.Errorf("Expected output to be: %#v, got: %#v", tc.output, gotOutput)
			}
		}

		byteBuf.Reset()
	}

}
