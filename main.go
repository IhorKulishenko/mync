package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func printUsage(writer io.Writer) {
	fmt.Fprintf(writer, "Usage: %s [cmd-a|cmd-b] -h", os.Args[0])
	fmt.Fprintln(writer)
	handleCmdA(writer, []string{"-h"})
	handleCmdB(writer, []string{"-h"})
}

func handleCmdA(writer io.Writer, args []string) error {
	var v string

	fs := flag.NewFlagSet("cmd-a", flag.ContinueOnError)
	fs.SetOutput(writer)
	fs.StringVar(&v, "verb", "argument-value", "Argument 1")

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	fmt.Fprint(writer, "Executing command A")

	return nil
}

func handleCmdB(writer io.Writer, args []string) error {
	var v string

	fs := flag.NewFlagSet("cmd-b", flag.ContinueOnError)
	fs.SetOutput(writer)
	fs.StringVar(&v, "verb", "argument-value", "Argument 1")

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	fmt.Fprint(writer, "Executing command B")

	return nil
}

func main() {
	var err error

	if len(os.Args) < 2 {
		printUsage(os.Stdout)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "cmd-a":
		err = handleCmdA(os.Stdout, os.Args[2:])

	case "cmd-b":
		err = handleCmdB(os.Stdout, os.Args[2:])

	default:
		printUsage(os.Stdout)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
