package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/marcgeld/cobrak/pkg/k8s"
	"github.com/marcgeld/cobrak/pkg/output"
	"github.com/marcgeld/cobrak/pkg/resources"
	"github.com/spf13/cobra"
)

func newResourcesUsageCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "usage",
		Short: "Show actual CPU/memory usage (requires metrics-server)",
		Long: `Displays actual CPU and memory usage per container using the metrics.k8s.io API.
Requires metrics-server to be installed in the cluster.`,
		RunE: runResourcesUsage,
	}

	addResourceFlags(c)

	return c
}

func runResourcesUsage(c *cobra.Command, _ []string) error {
	kubeconfig, _ := c.Root().PersistentFlags().GetString("kubeconfig")
	kubeCtx, _ := c.Root().PersistentFlags().GetString("context")
	namespace, _ := c.Flags().GetString("namespace")
	top, _ := c.Flags().GetInt("top")

	cfg, err := k8s.NewRestConfig(kubeconfig, kubeCtx)
	if err != nil {
		return fmt.Errorf("building rest config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	metricsReader, err := resources.NewMetricsReaderFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("building metrics client: %w", err)
	}

	available, err := metricsReader.IsAvailable(ctx)
	if err != nil {
		return fmt.Errorf("checking metrics availability: %w", err)
	}
	if !available {
		return fmt.Errorf("metrics API (metrics.k8s.io) not available; install metrics-server")
	}

	usages, err := metricsReader.PodMetrics(ctx, namespace)
	if err != nil {
		return fmt.Errorf("fetching pod metrics: %w", err)
	}

	w := c.OutOrStdout()
	fmt.Fprintln(w, output.RenderUsageTable(usages, top))

	return nil
}
