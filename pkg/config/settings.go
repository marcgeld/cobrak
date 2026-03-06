package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// PressureThresholds defines the pressure level thresholds
type PressureThresholds struct {
	Low       float64 `toml:"low"`
	Medium    float64 `toml:"medium"`
	High      float64 `toml:"high"`
	Saturated float64 `toml:"saturated"`
}

// Settings represents the cobrak configuration
type Settings struct {
	Output             string             `toml:"output"`
	Namespace          string             `toml:"namespace"`
	Context            string             `toml:"context"`
	Top                int                `toml:"top"`
	Color              bool               `toml:"color"`
	PressureThresholds PressureThresholds `toml:"pressure_thresholds"`
}

// DefaultSettings returns the default configuration
func DefaultSettings() *Settings {
	return &Settings{
		Output:    "text",
		Namespace: "",
		Context:   "",
		Top:       20,
		Color:     true,
		PressureThresholds: PressureThresholds{
			Low:       50.0,
			Medium:    75.0,
			High:      90.0,
			Saturated: 100.0,
		},
	}
}

// Validate checks if the pressure thresholds are in valid order
func (pt *PressureThresholds) Validate() error {
	if pt.Low < 0 || pt.Low > 100 {
		return fmt.Errorf("pressure threshold 'low' must be between 0 and 100, got %.1f", pt.Low)
	}
	if pt.Medium < 0 || pt.Medium > 100 {
		return fmt.Errorf("pressure threshold 'medium' must be between 0 and 100, got %.1f", pt.Medium)
	}
	if pt.High < 0 || pt.High > 100 {
		return fmt.Errorf("pressure threshold 'high' must be between 0 and 100, got %.1f", pt.High)
	}
	if pt.Saturated < 0 || pt.Saturated > 100 {
		return fmt.Errorf("pressure threshold 'saturated' must be between 0 and 100, got %.1f", pt.Saturated)
	}

	// Validate ordering: low < medium < high < saturated
	if pt.Low >= pt.Medium {
		return fmt.Errorf("pressure threshold 'low' (%.1f) must be less than 'medium' (%.1f)", pt.Low, pt.Medium)
	}
	if pt.Medium >= pt.High {
		return fmt.Errorf("pressure threshold 'medium' (%.1f) must be less than 'high' (%.1f)", pt.Medium, pt.High)
	}
	if pt.High >= pt.Saturated {
		return fmt.Errorf("pressure threshold 'high' (%.1f) must be less than 'saturated' (%.1f)", pt.High, pt.Saturated)
	}

	return nil
}

// LoadSettingsAt loads configuration from the given absolute path.
// If the file does not exist, default settings are returned.
func LoadSettingsAt(configPath string) (*Settings, error) {
	settings := DefaultSettings()

	// If config file doesn't exist, return defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return settings, nil
	}

	// Read and parse the config file
	_, err := toml.DecodeFile(configPath, settings)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", configPath, err)
	}

	// Validate pressure thresholds
	if err := settings.PressureThresholds.Validate(); err != nil {
		return nil, fmt.Errorf("invalid pressure thresholds in config: %w", err)
	}

	return settings, nil
}

// LoadSettings loads configuration using the resolved config path.
// Path precedence: --config flag (via COBRAK_CONFIG env) > COBRAK_CONFIG env > ~/.cobrak/settings.toml.
// Use LoadSettingsAt to specify an explicit resolved path.
func LoadSettings() (*Settings, error) {
	configPath, err := ResolveConfigPath("")
	if err != nil {
		// If we can't determine the path, fall back to defaults
		return DefaultSettings(), nil
	}
	return LoadSettingsAt(configPath)
}

// SaveSettingsAt saves configuration to the given absolute path.
func SaveSettingsAt(configPath string, settings *Settings) error {
	// Validate pressure thresholds before saving
	if err := settings.PressureThresholds.Validate(); err != nil {
		return err
	}

	// Create directory if it doesn't exist (private to the user)
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Use os.OpenRoot to scope file creation to configDir, preventing directory
	// traversal and satisfying gosec G304.
	dirRoot, err := os.OpenRoot(configDir)
	if err != nil {
		return fmt.Errorf("opening config directory: %w", err)
	}
	defer dirRoot.Close()

	file, err := dirRoot.OpenFile(filepath.Base(configPath), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(settings); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// SaveSettings saves configuration using the resolved config path.
// Path precedence: COBRAK_CONFIG env > ~/.cobrak/settings.toml.
// Use SaveSettingsAt to specify an explicit resolved path.
func SaveSettings(settings *Settings) error {
	configPath, err := ResolveConfigPath("")
	if err != nil {
		return fmt.Errorf("resolving config path: %w", err)
	}
	return SaveSettingsAt(configPath, settings)
}

// GetConfigPath returns the resolved configuration file path.
// Path precedence: COBRAK_CONFIG env > ~/.cobrak/settings.toml.
func GetConfigPath() (string, error) {
	return ResolveConfigPath("")
}

// FlagOverrides merges config file settings with command-line flags
// Command-line flags take precedence over config file settings
type FlagOverrides struct {
	Output    *string // nil means not provided via flag
	Namespace *string
	Context   *string
	Top       *int
}

// Merge applies flag overrides to config settings
func (s *Settings) Merge(overrides FlagOverrides) {
	if overrides.Output != nil {
		s.Output = *overrides.Output
	}
	if overrides.Namespace != nil {
		s.Namespace = *overrides.Namespace
	}
	if overrides.Context != nil {
		s.Context = *overrides.Context
	}
	if overrides.Top != nil {
		s.Top = *overrides.Top
	}
}
