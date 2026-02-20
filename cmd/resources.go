package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/marcgeld/cobrak/pkg/k8s"
	"github.com/marcgeld/cobrak/pkg/resources"
)

func newResourcesCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "resources",
		Short: "Broad summary of cluster resource usage",
		Long:  "Provides a high-level report combining inventory and policy summary, and indicates whether metrics are available.",
		RunE:  runResources,
	}

	addResourceFlags(c)

	c.AddCommand(newResourcesInventoryCmd())
	c.AddCommand(newResourcesUsageCmd())
	c.AddCommand(newResourcesDiffCmd())

	return c
}

func addResourceFlags(c *cobra.Command) {
	c.Flags().String("namespace", "", "namespace to inspect (default: all namespaces)")
	c.Flags().Bool("all-namespaces", true, "inspect all namespaces (default when --namespace is empty)")
	c.Flags().Int("top", 20, "number of top offenders to show")
	c.Flags().String("output", "table", "output format: table or json")
}

func runResources(c *cobra.Command, _ []string) error {
	kubeconfig, _ := c.Root().PersistentFlags().GetString("kubeconfig")
	kubeCtx, _ := c.Root().PersistentFlags().GetString("context")
	namespace, _ := c.Flags().GetString("namespace")

	cfg, err := k8s.NewRestConfig(kubeconfig, kubeCtx)
	if err != nil {
		return fmt.Errorf("building rest config: %w", err)
	}

	client, err := k8s.NewClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("building k8s client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	nsInventories, containers, policies, err := resources.BuildInventory(ctx, client, namespace)
	if err != nil {
		return fmt.Errorf("building inventory: %w", err)
	}

	totalContainers := 0
	missingRequests := 0
	missingLimits := 0
	for _, ns := range nsInventories {
		totalContainers += ns.ContainersTotal
		missingRequests += ns.ContainersMissingAnyRequests
		missingLimits += ns.ContainersMissingAnyLimits
	}
	_ = containers
	_ = policies

	fmt.Fprintf(c.OutOrStdout(), "Namespaces:                  %d\n", len(nsInventories))
	fmt.Fprintf(c.OutOrStdout(), "Total containers:            %d\n", totalContainers)
	fmt.Fprintf(c.OutOrStdout(), "Missing any requests:        %d\n", missingRequests)
	fmt.Fprintf(c.OutOrStdout(), "Missing any limits:          %d\n", missingLimits)

	metricsReader, err := resources.NewMetricsReaderFromConfig(cfg)
	if err == nil {
		available, _ := metricsReader.IsAvailable(ctx)
		if available {
			fmt.Fprintf(c.OutOrStdout(), "Metrics API:                 available\n")
		} else {
			fmt.Fprintf(c.OutOrStdout(), "Metrics API:                 not available (install metrics-server for usage data)\n")
		}
	}

	return nil
}
