package cmd

import (
	"fmt"

	"github.com/marcgeld/cobrak/pkg/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "config",
		Short: "Manage cobrak configuration",
		Long:  "Manage cobrak settings stored in ~/.cobrak/settings.toml",
	}

	c.AddCommand(newConfigSetCmd())
	c.AddCommand(newConfigShowCmd())
	c.AddCommand(newConfigResetCmd())

	return c
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: "Set configuration value",
		Long:  "Set a configuration value in ~/.cobrak/settings.toml",
		Args:  cobra.ExactArgs(2),
		RunE:  runConfigSet,
	}
}

func runConfigSet(c *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Load current settings
	settings, err := config.LoadSettings()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Update the setting based on key
	switch key {
	case "output":
		settings.Output = value
	case "namespace":
		settings.Namespace = value
	case "context":
		settings.Context = value
	case "top":
		var topVal int
		_, err := fmt.Sscanf(value, "%d", &topVal)
		if err != nil {
			return fmt.Errorf("invalid value for 'top': must be a number")
		}
		settings.Top = topVal
	case "pressure_thresholds.low":
		var val float64
		_, err := fmt.Sscanf(value, "%f", &val)
		if err != nil {
			return fmt.Errorf("invalid value for pressure threshold: must be a number")
		}
		settings.PressureThresholds.Low = val
	case "pressure_thresholds.medium":
		var val float64
		_, err := fmt.Sscanf(value, "%f", &val)
		if err != nil {
			return fmt.Errorf("invalid value for pressure threshold: must be a number")
		}
		settings.PressureThresholds.Medium = val
	case "pressure_thresholds.high":
		var val float64
		_, err := fmt.Sscanf(value, "%f", &val)
		if err != nil {
			return fmt.Errorf("invalid value for pressure threshold: must be a number")
		}
		settings.PressureThresholds.High = val
	case "pressure_thresholds.saturated":
		var val float64
		_, err := fmt.Sscanf(value, "%f", &val)
		if err != nil {
			return fmt.Errorf("invalid value for pressure threshold: must be a number")
		}
		settings.PressureThresholds.Saturated = val
	case "color":
		colorVal := value == "true" || value == "1" || value == "yes"
		settings.Color = colorVal
	default:
		return fmt.Errorf("unknown config key: %s (valid keys: output, namespace, context, top, color, pressure_thresholds.low, pressure_thresholds.medium, pressure_thresholds.high, pressure_thresholds.saturated)", key)
	}

	// Save settings
	if err := config.SaveSettings(settings); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	configPath, _ := config.GetConfigPath()
	fmt.Fprintf(c.OutOrStdout(), "✓ Configuration updated\n")
	fmt.Fprintf(c.OutOrStdout(), "  Config file: %s\n", configPath)
	fmt.Fprintf(c.OutOrStdout(), "  %s = %s\n", key, value)

	return nil
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  "Display the current configuration from ~/.cobrak/settings.toml",
		RunE:  runConfigShow,
	}
}

func runConfigShow(c *cobra.Command, _ []string) error {
	// Load settings
	settings, err := config.LoadSettings()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	configPath, _ := config.GetConfigPath()

	fmt.Fprintf(c.OutOrStdout(), "Configuration file: %s\n\n", configPath)
	fmt.Fprintf(c.OutOrStdout(), "output:    %s (text, json, yaml)\n", settings.Output)
	fmt.Fprintf(c.OutOrStdout(), "namespace: %s (empty = all namespaces)\n", settings.Namespace)
	fmt.Fprintf(c.OutOrStdout(), "context:   %s (empty = current context)\n", settings.Context)
	fmt.Fprintf(c.OutOrStdout(), "top:       %d\n", settings.Top)
	colorStatus := "enabled"
	if !settings.Color {
		colorStatus = "disabled"
	}
	fmt.Fprintf(c.OutOrStdout(), "color:     %s (true or false)\n", colorStatus)
	fmt.Fprintf(c.OutOrStdout(), "\nPressure Thresholds:\n")
	fmt.Fprintf(c.OutOrStdout(), "  low:       %.1f (0-100)\n", settings.PressureThresholds.Low)
	fmt.Fprintf(c.OutOrStdout(), "  medium:    %.1f (0-100, must be > low)\n", settings.PressureThresholds.Medium)
	fmt.Fprintf(c.OutOrStdout(), "  high:      %.1f (0-100, must be > medium)\n", settings.PressureThresholds.High)
	fmt.Fprintf(c.OutOrStdout(), "  saturated: %.1f (0-100, must be > high)\n", settings.PressureThresholds.Saturated)
	fmt.Fprintf(c.OutOrStdout(), "\nNote: Command-line flags (like --nocolor) override these settings\n")

	return nil
}

func newConfigResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		Long:  "Reset all configuration settings to their default values",
		RunE:  runConfigReset,
	}
}

func runConfigReset(c *cobra.Command, _ []string) error {
	// Create default settings
	settings := config.DefaultSettings()

	// Save settings
	if err := config.SaveSettings(settings); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Fprintf(c.OutOrStdout(), "✓ Configuration reset to defaults\n")
	fmt.Fprintf(c.OutOrStdout(), "  output:    %s\n", settings.Output)
	fmt.Fprintf(c.OutOrStdout(), "  namespace: %s\n", settings.Namespace)
	fmt.Fprintf(c.OutOrStdout(), "  context:   %s\n", settings.Context)
	fmt.Fprintf(c.OutOrStdout(), "  top:       %d\n", settings.Top)

	return nil
}
