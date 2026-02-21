package main

import (
	"fmt"
	"os"

	"github.com/marcgeld/cobrak/cmd"
)

// Version variables set at build time
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Pass version info to commands
	cmd.SetVersion(version, commit, date)

	root := cmd.NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
