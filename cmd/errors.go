package cmd

import "errors"

var ErrNoServerSpecified = errors.New("you have to specify the remote server")
var ErrInvalidHttpMethod = errors.New("invalid HTTP method")
var ErrInvalidHttpUsage = errors.New("invalid HTTP subcommand usage")
