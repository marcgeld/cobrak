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

func newResourcesInventoryCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "inventory",
		Short: "Show pod/container resource requests/limits coverage",
		Long: `Displays per-namespace totals for CPU/memory requests and limits,
highlights containers missing requests/limits, and shows LimitRange/ResourceQuota summaries.`,
		RunE: runResourcesInventory,
	}

	addResourceFlags(c)

	return c
}

func runResourcesInventory(c *cobra.Command, _ []string) error {
	kubeconfig, _ := c.Root().PersistentFlags().GetString("kubeconfig")
	kubeCtx, _ := c.Root().PersistentFlags().GetString("context")
	namespace, _ := c.Flags().GetString("namespace")
	top, _ := c.Flags().GetInt("top")

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

	w := c.OutOrStdout()

	fmt.Fprintln(w, output.RenderNamespaceInventoryTable(nsInventories))
	if top > 0 {
		fmt.Fprintln(w, output.RenderMissingResourcesTable(containers, top))
	}
	fmt.Fprintln(w, output.RenderPolicySummary(policies))

	return nil
}
