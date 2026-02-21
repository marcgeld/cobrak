package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display version, commit hash, and build date information.",
		Run:   runVersion,
	}
}

func runVersion(c *cobra.Command, _ []string) {
	v, commit, date := GetVersion()
	fmt.Fprintf(c.OutOrStdout(), "cobrak version %s\n", v)
	fmt.Fprintf(c.OutOrStdout(), "commit: %s\n", commit)
	fmt.Fprintf(c.OutOrStdout(), "date: %s\n", date)
}
