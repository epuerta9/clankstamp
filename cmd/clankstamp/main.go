// Command clankstamp is the CLI entry point for the clankstamp replay toolchain.
//
// See PRD.md for the full product spec. This is the bootstrap entry point — it
// wires up subcommand dispatch and delegates to internal/cli.
package main

import (
	"fmt"
	"os"

	"github.com/epuerta9/clankstamp/internal/cli"
)

// version is overridden at build time via -ldflags.
var version = "0.0.0-dev"

func main() {
	if err := cli.Run(os.Args[1:], version); err != nil {
		fmt.Fprintln(os.Stderr, "clankstamp:", err)
		os.Exit(1)
	}
}
