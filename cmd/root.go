package cmd

import (
	"github.com/spf13/cobra"
)

var kubeconfigFlag string

// NewRootCmd returns the root cobra command for cobrak.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "cobrak",
		Short: "cobrak is a modular Kubernetes cluster analysis tool",
		Long:  "cobrak inspects cluster state from the command line. Use a subcommand to analyse a specific aspect.",
	}

	root.PersistentFlags().StringVar(&kubeconfigFlag, "kubeconfig", "", "path to kubeconfig file (overrides KUBECONFIG env and ~/.kube/config)")

	root.AddCommand(newCapacityCmd(&kubeconfigFlag))

	return root
}
