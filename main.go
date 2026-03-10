package main

import (
	"context"
	"os"

	"github.com/gaebalai/secfarm/pkg/cli"
)

func main() {
	ctx := context.Background()
	if err := cli.Run(ctx, os.Args); err != nil {
		os.Exit(err.Code)
	}
}
