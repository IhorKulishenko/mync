package cmd

import (
	"flag"
	"fmt"
	"io"
)

type grpcConfig struct {
	server string
	method string
	body   string
}

func HandleGrpc(writer io.Writer, args []string) error {
	c := grpcConfig{}

	fs := flag.NewFlagSet("grpc", flag.ContinueOnError)
	fs.SetOutput(writer)
	fs.StringVar(&c.method, "method", "", "Method to call")
	fs.StringVar(&c.body, "body", "", "Body of request")
	fs.Usage = func() {
		var usageString = `
grpc: A gRPC client.
grpc: <options> server`

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

	c.server = fs.Arg(0)
	fmt.Fprintln(writer, "Executing grpc command")

	return nil
}
