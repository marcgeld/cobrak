package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/marcgeld/cobrak/pkg/capacity"
	"github.com/marcgeld/cobrak/pkg/config"
	"github.com/marcgeld/cobrak/pkg/k8s"
	"github.com/marcgeld/cobrak/pkg/output"
	"github.com/marcgeld/cobrak/pkg/resources"
	"github.com/spf13/cobra"
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
	c.Flags().String("output", "text", "output format: text, json, or yaml")
}

func runResources(c *cobra.Command, _ []string) error {
	// Load configuration from ~/.cobrak/settings.toml
	settings, err := config.LoadSettings()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Get flag values (may be empty/zero)
	kubeconfig, _ := c.Root().PersistentFlags().GetString("kubeconfig")
	kubeCtx, _ := c.Root().PersistentFlags().GetString("context")

	// Get resource-specific flags
	flagNamespace, _ := c.Flags().GetString("namespace")
	flagTop, _ := c.Flags().GetInt("top")
	flagOutput, _ := c.Flags().GetString("output")

	// Determine if flags were explicitly set (not just default values)
	// We check if the flag was actually provided on the command line
	outputFlagSet := c.Flag("output").Changed
	namespaceFlagSet := c.Flag("namespace").Changed
	topFlagSet := c.Flag("top").Changed

	// Build flag overrides (only set if flag was explicitly provided)
	overrides := config.FlagOverrides{}
	if outputFlagSet {
		overrides.Output = &flagOutput
	}
	if namespaceFlagSet {
		overrides.Namespace = &flagNamespace
	}
	if topFlagSet {
		overrides.Top = &flagTop
	}

	// Merge config with flags (flags take precedence)
	settings.Merge(overrides)

	// Use merged settings
	namespace := settings.Namespace
	outputFormat := settings.Output

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

	// Get cluster capacity summary
	summary, err := capacity.AnalyzeSummary(ctx, client, namespace)
	if err != nil {
		return fmt.Errorf("analyzing capacity summary: %w", err)
	}

	// Get pod-level resource summaries
	podSummaries, err := resources.BuildPodSummaries(ctx, client, namespace)
	if err != nil {
		return fmt.Errorf("building pod summaries: %w", err)
	}

	// Get inventory
	nsInventories, containers, policies, err := resources.BuildInventory(ctx, client, namespace)
	if err != nil {
		return fmt.Errorf("building inventory: %w", err)
	}

	_ = containers
	_ = policies

	// Check metrics availability
	metricsAvailable := false
	metricsReader, err := resources.NewMetricsReaderFromConfig(cfg)
	if err == nil {
		available, _ := metricsReader.IsAvailable(ctx)
		metricsAvailable = available
	}

	// Parse output format
	format, err := output.ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// For text format, use the original text output
	if format == output.FormatText {
		fmt.Fprintf(c.OutOrStdout(), "\n=== CLUSTER CAPACITY SUMMARY ===\n")
		fmt.Fprintf(c.OutOrStdout(), "CPU Capacity:          %s\n", summary.TotalCPUCapacity.String())
		fmt.Fprintf(c.OutOrStdout(), "CPU Allocatable:       %s\n", summary.TotalCPUAllocatable.String())
		fmt.Fprintf(c.OutOrStdout(), "CPU Requests:          %s\n", summary.TotalCPURequests.String())
		fmt.Fprintf(c.OutOrStdout(), "CPU Limits:            %s\n", summary.TotalCPULimits.String())
		fmt.Fprintf(c.OutOrStdout(), "\nMemory Capacity:       %s\n", summary.TotalMemCapacity.String())
		fmt.Fprintf(c.OutOrStdout(), "Memory Allocatable:    %s\n", summary.TotalMemAllocatable.String())
		fmt.Fprintf(c.OutOrStdout(), "Memory Requests:       %s\n", summary.TotalMemRequests.String())
		fmt.Fprintf(c.OutOrStdout(), "Memory Limits:         %s\n", summary.TotalMemLimits.String())

		fmt.Fprintf(c.OutOrStdout(), "\n=== POD RESOURCE DETAILS ===\n")
		if len(podSummaries) > 0 {
			fmt.Fprintf(c.OutOrStdout(), "%s\n\n", output.RenderPodResourceSummary(podSummaries))
			fmt.Fprintf(c.OutOrStdout(), "%s\n", output.RenderPodResourceSummaryTotals(podSummaries))
		} else {
			fmt.Fprintf(c.OutOrStdout(), "No pods found.\n")
		}

		totalContainers := 0
		missingRequests := 0
		missingLimits := 0
		for _, ns := range nsInventories {
			totalContainers += ns.ContainersTotal
			missingRequests += ns.ContainersMissingAnyRequests
			missingLimits += ns.ContainersMissingAnyLimits
		}

		fmt.Fprintf(c.OutOrStdout(), "\n=== RESOURCE INVENTORY ===\n")
		fmt.Fprintf(c.OutOrStdout(), "Namespaces:                  %d\n", len(nsInventories))
		fmt.Fprintf(c.OutOrStdout(), "Total containers:            %d\n", totalContainers)
		fmt.Fprintf(c.OutOrStdout(), "Missing any requests:        %d\n", missingRequests)
		fmt.Fprintf(c.OutOrStdout(), "Missing any limits:          %d\n", missingLimits)

		if metricsAvailable {
			fmt.Fprintf(c.OutOrStdout(), "Metrics API:                 available\n")
		} else {
			fmt.Fprintf(c.OutOrStdout(), "Metrics API:                 not available (install metrics-server for usage data)\n")
		}

		return nil
	}

	// For JSON/YAML formats, create structured output
	resourcesSummary := buildResourcesSummary(summary, podSummaries, nsInventories, metricsAvailable)

	outputStr, err := output.RenderOutput(resourcesSummary, format)
	if err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	fmt.Fprintf(c.OutOrStdout(), "%s\n", outputStr)
	return nil
}

// buildResourcesSummary creates a structured summary for JSON/YAML output
func buildResourcesSummary(
	summary *capacity.ClusterCapacitySummary,
	podSummaries []resources.PodResourceSummary,
	nsInventories []resources.NamespaceInventory,
	metricsAvailable bool,
) *output.ResourcesSummary {
	// Build cluster capacity
	clusterCap := &output.ClusterCapacitySummary{
		CPUCapacity:    summary.TotalCPUCapacity.String(),
		CPUAllocatable: summary.TotalCPUAllocatable.String(),
		CPURequests:    summary.TotalCPURequests.String(),
		CPULimits:      summary.TotalCPULimits.String(),
		MemCapacity:    summary.TotalMemCapacity.String(),
		MemAllocatable: summary.TotalMemAllocatable.String(),
		MemRequests:    summary.TotalMemRequests.String(),
		MemLimits:      summary.TotalMemLimits.String(),
	}

	// Build pod details
	podDetails := make([]output.PodDetail, len(podSummaries))
	for i, pod := range podSummaries {
		podDetails[i] = output.PodDetail{
			Namespace:  pod.Namespace,
			Pod:        pod.PodName,
			CPURequest: pod.CPURequest.String(),
			CPULimit:   pod.CPULimit.String(),
			MemRequest: pod.MemRequest.String(),
			MemLimit:   pod.MemLimit.String(),
		}
	}

	// Build namespace inventory
	nsInv := make([]output.NamespaceSummary, len(nsInventories))
	for i, ns := range nsInventories {
		nsInv[i] = output.NamespaceSummary{
			Namespace:       ns.Namespace,
			ContainersTotal: ns.ContainersTotal,
			MissingRequests: ns.ContainersMissingAnyRequests,
			MissingLimits:   ns.ContainersMissingAnyLimits,
			CPURequests:     ns.CPURequestsTotal.String(),
			CPULimits:       ns.CPULimitsTotal.String(),
			MemRequests:     ns.MemRequestsTotal.String(),
			MemLimits:       ns.MemLimitsTotal.String(),
		}
	}

	return &output.ResourcesSummary{
		ClusterCapacity:    clusterCap,
		PodDetails:         podDetails,
		NamespaceInventory: nsInv,
		MetricsAvailable:   metricsAvailable,
	}
}

func newResourcesSimpleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "simple",
		Short: "Quick cluster resource pressure summary",
		Long:  "Shows a simple one-liner summary of cluster pressure and resource constraints per node and namespace.",
		RunE:  runResourcesSimple,
	}

	return nil
}
