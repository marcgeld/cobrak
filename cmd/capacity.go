package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/marcgeld/cobrak/pkg/capacity"
	"github.com/marcgeld/cobrak/pkg/k8s"
	"github.com/marcgeld/cobrak/pkg/kubeconfig"
)

func newCapacityCmd(kubeconfigFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:   "capacity",
		Short: "Show CPU and memory capacity for each node",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolver := kubeconfig.NewDefaultResolver()
			kubeconfigPath, err := resolver.Resolve(*kubeconfigFlag)
			if err != nil {
				return fmt.Errorf("resolving kubeconfig: %w", err)
			}

			client, err := k8s.NewClient(kubeconfigPath)
			if err != nil {
				return fmt.Errorf("creating k8s client: %w", err)
			}

			nodes, err := capacity.Analyze(context.Background(), client)
			if err != nil {
				return fmt.Errorf("analysing capacity: %w", err)
			}

			for _, n := range nodes {
				fmt.Fprintf(cmd.OutOrStdout(), "Node: %s\n", n.Name)
				fmt.Fprintf(cmd.OutOrStdout(), "CPU: %s alloc / %s cap\n",
					n.CPUAllocatable.String(), n.CPUCapacity.String())
				fmt.Fprintf(cmd.OutOrStdout(), "Memory: %s alloc / %s cap\n\n",
					n.MemAllocatable.String(), n.MemCapacity.String())
			}

			return nil
		},
	}
}
