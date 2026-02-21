package cmd

import (
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	kubeconfig := ""

	root := &cobra.Command{
		Use:   "cobrak",
		Short: "cobrak - analytical CLI for Kubernetes cluster state",
		Long:  "cobrak is a modular, lightweight, fast analytical tool for inspecting cluster state from the command line.",
	}

	root.PersistentFlags().String("kubeconfig", "", "path to kubeconfig file (default: KUBECONFIG env or ~/.kube/config)")
	root.PersistentFlags().String("context", "", "kubeconfig context to use")

	root.AddCommand(newResourcesCmd())
	root.AddCommand(newCapacityCmd(&kubeconfig))
	root.AddCommand(newNodeInfoCmd())
	root.AddCommand(newConfigCmd())

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
