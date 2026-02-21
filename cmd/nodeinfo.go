package cmd

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/marcgeld/cobrak/pkg/k8s"
	"github.com/marcgeld/cobrak/pkg/nodeinfo"
	"github.com/spf13/cobra"
)

func newNodeInfoCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "nodeinfo",
		Short: "Detailed node system information",
		Long:  "Shows OS, kernel, CPU, GPU, memory pressure, filesystem latency, container runtime, and virtualization details for nodes.",
		RunE:  runNodeInfo,
	}

	c.Flags().String("node", "", "specific node name (default: all nodes)")
	c.Flags().Bool("compact", false, "show compact format")
	c.Flags().Bool("health", false, "show only health status")

	return c
}

func runNodeInfo(c *cobra.Command, _ []string) error {
	kubeconfig, _ := c.Root().PersistentFlags().GetString("kubeconfig")
	kubeCtx, _ := c.Root().PersistentFlags().GetString("context")
	nodeName, _ := c.Flags().GetString("node")
	compact, _ := c.Flags().GetBool("compact")
	healthOnly, _ := c.Flags().GetBool("health")

	cfg, err := k8s.NewRestConfig(kubeconfig, kubeCtx)
	if err != nil {
		return fmt.Errorf("building rest config: %w", err)
	}

	client, err := k8s.NewClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("building k8s client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Analyze specific node or all nodes
	if nodeName != "" {
		info, err := nodeinfo.AnalyzeNode(ctx, client, nodeName)
		if err != nil {
			return fmt.Errorf("analyzing node %s: %w", nodeName, err)
		}

		if healthOnly {
			health, err := nodeinfo.GetNodeHealthStatus(ctx, client, nodeName)
			if err != nil {
				return fmt.Errorf("getting node health: %w", err)
			}
			fmt.Fprintf(c.OutOrStdout(), "%s\n", nodeinfo.RenderNodeHealth(health))
		} else if compact {
			fmt.Fprintf(c.OutOrStdout(), "%s\n", nodeinfo.RenderNodeInfoCompact(info))
		} else {
			fmt.Fprintf(c.OutOrStdout(), "%s\n", nodeinfo.RenderNodeInfo(info))
		}
	} else {
		// Analyze all nodes
		infos, err := nodeinfo.AnalyzeAllNodes(ctx, client)
		if err != nil {
			return fmt.Errorf("analyzing all nodes: %w", err)
		}

		// Sort by node name
		sort.Slice(infos, func(i, j int) bool {
			return infos[i].NodeName < infos[j].NodeName
		})

		if healthOnly {
			// Show health status for all nodes
			fmt.Fprintf(c.OutOrStdout(), "=== NODE HEALTH STATUS ===\n\n")
			for _, info := range infos {
				health, err := nodeinfo.GetNodeHealthStatus(ctx, client, info.NodeName)
				if err != nil {
					continue
				}
				fmt.Fprintf(c.OutOrStdout(), "%s\n\n", nodeinfo.RenderNodeHealth(health))
			}
		} else if compact {
			fmt.Fprintf(c.OutOrStdout(), "%s\n", nodeinfo.RenderMultipleNodeInfoCompact(infos))
		} else {
			// Show detailed info for all nodes
			for i, info := range infos {
				fmt.Fprintf(c.OutOrStdout(), "%s\n", nodeinfo.RenderNodeInfo(&info))
				if i < len(infos)-1 {
					fmt.Fprintf(c.OutOrStdout(), "\n---\n\n")
				}
			}
		}
	}

	return nil
}
