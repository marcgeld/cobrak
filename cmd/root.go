package cmd

import (
	"github.com/spf13/cobra"
)

// Version information
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// SetVersion sets the version information
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
}

// GetVersion returns the version information
func GetVersion() (string, string, string) {
	return version, commit, date
}

func newRootCmd() *cobra.Command {
	kubeconfig := ""

	root := &cobra.Command{
		Use:   "cobrak",
		Short: "cobrak - analytical CLI for Kubernetes cluster state",
		Long:  "cobrak is a modular, lightweight, fast analytical tool for inspecting cluster state from the command line.",
	}

	root.PersistentFlags().String("kubeconfig", "", "path to kubeconfig file (default: KUBECONFIG env or ~/.kube/config)")
	root.PersistentFlags().String("context", "", "kubeconfig context to use")
	root.PersistentFlags().Bool("nocolor", false, "disable colored output")

	root.AddCommand(newResourcesCmd())
	root.AddCommand(newCapacityCmd(&kubeconfig))
	root.AddCommand(newNodeInfoCmd())
	root.AddCommand(newConfigCmd())
	root.AddCommand(newVersionCmd())

	return root
}

// NewRootCmd returns the root command for cobrak.
func NewRootCmd() *cobra.Command {
	return newRootCmd()
}

// Execute runs the root command.
func Execute() error {
	return newRootCmd().Execute()
}
