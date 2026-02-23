package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/marcgeld/cobrak/pkg/config"
	"github.com/marcgeld/cobrak/pkg/k8s"
	"github.com/marcgeld/cobrak/pkg/output"
	"github.com/marcgeld/cobrak/pkg/resources"
	"github.com/spf13/cobra"
)

func newResourcesDiffCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "diff",
		Short: "Compare actual usage vs requests/limits (requires metrics-server)",
		Long: `Compares actual CPU/memory usage against requested resources to identify:
  - Waste candidates: usage much lower than requests
  - Pressure candidates: usage higher than or close to requests/limits
Requires metrics-server to be installed in the cluster.`,
		RunE: runResourcesDiff,
	}

	addResourceFlags(c)

	return c
}

func runResourcesDiff(c *cobra.Command, _ []string) error {
	kubeconfig, _ := c.Root().PersistentFlags().GetString("kubeconfig")
	kubeCtx, _ := c.Root().PersistentFlags().GetString("context")
	nocolor, _ := c.Root().PersistentFlags().GetBool("nocolor")
	namespace, _ := c.Flags().GetString("namespace")
	top, _ := c.Flags().GetInt("top")

	// Load configuration and set color
	settings, err := config.LoadSettings()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	colorEnabled := settings.Color && !nocolor
	output.SetGlobalColorEnabled(colorEnabled)

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

	_, containers, _, err := resources.BuildInventory(ctx, client, namespace)
	if err != nil {
		return fmt.Errorf("building inventory: %w", err)
	}

	usages, err := metricsReader.PodMetrics(ctx, namespace)
	if err != nil {
		return fmt.Errorf("fetching pod metrics: %w", err)
	}

	diffs := resources.BuildDiff(containers, usages)

	w := c.OutOrStdout()
	fmt.Fprintln(w, output.RenderDiffTable(diffs, top))

	return nil
}
