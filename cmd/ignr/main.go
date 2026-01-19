package main

import (
	"go.seanlatimer.dev/ignr/pkg/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		cli.ExitWithError(err)
	}
}
