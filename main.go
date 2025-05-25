package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"example.com/sub-cmd-example/cmd"
)

var errInvalidSubCommand = errors.New("invalid sub-command specified")

func printUsage(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: mync [http|grpc] -h")
	cmd.HandleHttp(writer, []string{"-h"})
	cmd.HandleGrpc(writer, []string{"-h"})
}

func handleCommand(writer io.Writer, args []string) error {
	var err error

	if len(args) < 1 {
		err = errInvalidSubCommand
	} else {
		switch args[0] {
		case "http":
			err = cmd.HandleHttp(writer, args[1:])
		case "grpc":
			err = cmd.HandleGrpc(writer, args[1:])
		case "-h":
			printUsage(writer)
		case "-help":
			printUsage(writer)
		default:
			err = errInvalidSubCommand
		}
	}

	if errors.Is(err, cmd.ErrNoServerSpecified) || errors.Is(err, errInvalidSubCommand) {
		fmt.Fprintln(writer, err)
		printUsage(writer)
	}

	return err
}

func main() {
	if err := handleCommand(os.Stdout, os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
