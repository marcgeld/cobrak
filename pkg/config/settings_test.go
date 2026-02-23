package config

import (
	"os"
	"path/filepath"
	"testing"
)

// ...existing code...

func TestLoadAndParseValidTOMLFile(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()

	// Set HOME to temp directory for this test
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Create a valid TOML file
	configDir := filepath.Join(tempDir, ".cobrak")
	os.MkdirAll(configDir, 0755)

	configPath := filepath.Join(configDir, "settings.toml")
	tomlContent := `output = "json"
namespace = "production"
context = "prod-cluster"
top = 50

[pressure_thresholds]
low = 40.0
medium = 65.0
high = 85.0
saturated = 100.0
`

	if err := os.WriteFile(configPath, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write test TOML file: %v", err)
	}

	// Load settings
	settings, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	// Verify all settings were parsed correctly
	if settings.Output != "json" {
		t.Errorf("expected output 'json', got '%s'", settings.Output)
	}
	if settings.Namespace != "production" {
		t.Errorf("expected namespace 'production', got '%s'", settings.Namespace)
	}
	if settings.Context != "prod-cluster" {
		t.Errorf("expected context 'prod-cluster', got '%s'", settings.Context)
	}
	if settings.Top != 50 {
		t.Errorf("expected top 50, got %d", settings.Top)
	}

	// Verify pressure thresholds were parsed
	if settings.PressureThresholds.Low != 40.0 {
		t.Errorf("expected Low 40.0, got %.1f", settings.PressureThresholds.Low)
	}
	if settings.PressureThresholds.Medium != 65.0 {
		t.Errorf("expected Medium 65.0, got %.1f", settings.PressureThresholds.Medium)
	}
	if settings.PressureThresholds.High != 85.0 {
		t.Errorf("expected High 85.0, got %.1f", settings.PressureThresholds.High)
	}
	if settings.PressureThresholds.Saturated != 100.0 {
		t.Errorf("expected Saturated 100.0, got %.1f", settings.PressureThresholds.Saturated)
	}
}

func TestLoadTOMLWithPartialPressureThresholds(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()

	// Set HOME to temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Create TOML file with only some settings
	configDir := filepath.Join(tempDir, ".cobrak")
	os.MkdirAll(configDir, 0755)

	configPath := filepath.Join(configDir, "settings.toml")
	tomlContent := `output = "yaml"

[pressure_thresholds]
low = 30.0
medium = 60.0
`

	if err := os.WriteFile(configPath, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write test TOML file: %v", err)
	}

	// Load settings - partial thresholds are OK because unspecified values use defaults
	// The merged result (low=30, medium=60, high=90 (default), saturated=100 (default)) is valid
	settings, err := LoadSettings()
	if err != nil {
		t.Errorf("expected LoadSettings to succeed with partial pressure thresholds, but got error: %v", err)
	}
	if settings.Output != "yaml" {
		t.Errorf("expected output to be 'yaml', got %s", settings.Output)
	}
	if settings.PressureThresholds.Low != 30.0 {
		t.Errorf("expected Low to be 30.0, got %.1f", settings.PressureThresholds.Low)
	}
	if settings.PressureThresholds.Medium != 60.0 {
		t.Errorf("expected Medium to be 60.0, got %.1f", settings.PressureThresholds.Medium)
	}
}

func TestLoadInvalidTOMLFile(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()

	// Set HOME to temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Create an invalid TOML file
	configDir := filepath.Join(tempDir, ".cobrak")
	os.MkdirAll(configDir, 0755)

	configPath := filepath.Join(configDir, "settings.toml")
	tomlContent := `this is not valid TOML
[unclosed section
output = json without quotes
`

	if err := os.WriteFile(configPath, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write test TOML file: %v", err)
	}

	// Load settings - should fail due to invalid TOML
	_, err := LoadSettings()
	if err == nil {
		t.Errorf("expected LoadSettings to fail with invalid TOML, but it succeeded")
	}
}

func TestLoadTOMLWithInvalidPressureThresholds(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()

	// Set HOME to temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Create TOML file with invalid pressure thresholds (medium >= high)
	configDir := filepath.Join(tempDir, ".cobrak")
	os.MkdirAll(configDir, 0755)

	configPath := filepath.Join(configDir, "settings.toml")
	tomlContent := `output = "text"

[pressure_thresholds]
low = 50.0
medium = 90.0
high = 90.0
saturated = 100.0
`

	if err := os.WriteFile(configPath, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write test TOML file: %v", err)
	}

	// Load settings - should fail validation
	_, err := LoadSettings()
	if err == nil {
		t.Errorf("expected LoadSettings to fail with invalid thresholds, but it succeeded")
	}
}

func TestSaveAndParseTOMLFile(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()

	// Set HOME to temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Create settings with custom pressure thresholds
	settings := &Settings{
		Output:    "yaml",
		Namespace: "kube-system",
		Context:   "test-cluster",
		Top:       100,
		PressureThresholds: PressureThresholds{
			Low:       45.0,
			Medium:    70.0,
			High:      88.0,
			Saturated: 99.0,
		},
	}

	// Save settings
	if err := SaveSettings(settings); err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	// Load the settings back
	loadedSettings, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	// Verify all settings match
	if loadedSettings.Output != settings.Output {
		t.Errorf("output mismatch: expected '%s', got '%s'", settings.Output, loadedSettings.Output)
	}
	if loadedSettings.Namespace != settings.Namespace {
		t.Errorf("namespace mismatch: expected '%s', got '%s'", settings.Namespace, loadedSettings.Namespace)
	}
	if loadedSettings.Context != settings.Context {
		t.Errorf("context mismatch: expected '%s', got '%s'", settings.Context, loadedSettings.Context)
	}
	if loadedSettings.Top != settings.Top {
		t.Errorf("top mismatch: expected %d, got %d", settings.Top, loadedSettings.Top)
	}

	// Verify pressure thresholds match
	if loadedSettings.PressureThresholds.Low != settings.PressureThresholds.Low {
		t.Errorf("Low mismatch: expected %.1f, got %.1f", settings.PressureThresholds.Low, loadedSettings.PressureThresholds.Low)
	}
	if loadedSettings.PressureThresholds.Medium != settings.PressureThresholds.Medium {
		t.Errorf("Medium mismatch: expected %.1f, got %.1f", settings.PressureThresholds.Medium, loadedSettings.PressureThresholds.Medium)
	}
	if loadedSettings.PressureThresholds.High != settings.PressureThresholds.High {
		t.Errorf("High mismatch: expected %.1f, got %.1f", settings.PressureThresholds.High, loadedSettings.PressureThresholds.High)
	}
	if loadedSettings.PressureThresholds.Saturated != settings.PressureThresholds.Saturated {
		t.Errorf("Saturated mismatch: expected %.1f, got %.1f", settings.PressureThresholds.Saturated, loadedSettings.PressureThresholds.Saturated)
	}
}

func TestTOMLFileRoundTrip(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()

	// Set HOME to temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Create initial settings
	originalSettings := &Settings{
		Output:    "json",
		Namespace: "default",
		Context:   "minikube",
		Top:       25,
		PressureThresholds: PressureThresholds{
			Low:       35.0,
			Medium:    60.0,
			High:      82.0,
			Saturated: 98.0,
		},
	}

	// Save settings
	if err := SaveSettings(originalSettings); err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	// Load settings
	loadedSettings, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	// Modify and save again
	loadedSettings.Output = "yaml"
	loadedSettings.Top = 30
	loadedSettings.PressureThresholds.Low = 40.0

	if err := SaveSettings(loadedSettings); err != nil {
		t.Fatalf("SaveSettings (second save) failed: %v", err)
	}

	// Load again
	finalSettings, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings (second load) failed: %v", err)
	}

	// Verify the modified settings
	if finalSettings.Output != "yaml" {
		t.Errorf("expected output 'yaml', got '%s'", finalSettings.Output)
	}
	if finalSettings.Top != 30 {
		t.Errorf("expected top 30, got %d", finalSettings.Top)
	}
	if finalSettings.PressureThresholds.Low != 40.0 {
		t.Errorf("expected Low 40.0, got %.1f", finalSettings.PressureThresholds.Low)
	}
}

func TestLoadTOMLWithEmptyPressureThresholds(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()

	// Set HOME to temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Create TOML file without pressure_thresholds section
	configDir := filepath.Join(tempDir, ".cobrak")
	os.MkdirAll(configDir, 0755)

	configPath := filepath.Join(configDir, "settings.toml")
	tomlContent := `output = "text"
namespace = "default"
context = ""
top = 20
`

	if err := os.WriteFile(configPath, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write test TOML file: %v", err)
	}

	// Load settings
	settings, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	// Verify that defaults were used for pressure thresholds
	if settings.PressureThresholds.Low != 50.0 {
		t.Errorf("expected default Low 50.0, got %.1f", settings.PressureThresholds.Low)
	}
	if settings.PressureThresholds.Medium != 75.0 {
		t.Errorf("expected default Medium 75.0, got %.1f", settings.PressureThresholds.Medium)
	}
	if settings.PressureThresholds.High != 90.0 {
		t.Errorf("expected default High 90.0, got %.1f", settings.PressureThresholds.High)
	}
	if settings.PressureThresholds.Saturated != 100.0 {
		t.Errorf("expected default Saturated 100.0, got %.1f", settings.PressureThresholds.Saturated)
	}
}

func TestPressureThresholdsValidation(t *testing.T) {
	tests := []struct {
		name       string
		thresholds PressureThresholds
		shouldFail bool
	}{
		{
			name: "valid thresholds in correct order",
			thresholds: PressureThresholds{
				Low:       50.0,
				Medium:    75.0,
				High:      90.0,
				Saturated: 100.0,
			},
			shouldFail: false,
		},
		{
			name: "valid custom thresholds",
			thresholds: PressureThresholds{
				Low:       40.0,
				Medium:    60.0,
				High:      80.0,
				Saturated: 95.0,
			},
			shouldFail: false,
		},
		{
			name: "medium >= high (invalid)",
			thresholds: PressureThresholds{
				Low:       50.0,
				Medium:    90.0,
				High:      90.0,
				Saturated: 100.0,
			},
			shouldFail: true,
		},
		{
			name: "low >= medium (invalid)",
			thresholds: PressureThresholds{
				Low:       75.0,
				Medium:    75.0,
				High:      90.0,
				Saturated: 100.0,
			},
			shouldFail: true,
		},
		{
			name: "high >= saturated (invalid)",
			thresholds: PressureThresholds{
				Low:       50.0,
				Medium:    75.0,
				High:      100.0,
				Saturated: 100.0,
			},
			shouldFail: true,
		},
		{
			name: "negative value (invalid)",
			thresholds: PressureThresholds{
				Low:       -10.0,
				Medium:    75.0,
				High:      90.0,
				Saturated: 100.0,
			},
			shouldFail: true,
		},
		{
			name: "value > 100 (invalid)",
			thresholds: PressureThresholds{
				Low:       50.0,
				Medium:    75.0,
				High:      90.0,
				Saturated: 150.0,
			},
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.thresholds.Validate()
			if tt.shouldFail && err == nil {
				t.Errorf("expected validation to fail, but it passed")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("expected validation to pass, but got error: %v", err)
			}
		})
	}
}

func TestMergeWithFlags(t *testing.T) {
	settings := &Settings{
		Output:    "text",
		Namespace: "default",
		Top:       20,
	}

	output := "json"
	top := 50

	overrides := FlagOverrides{
		Output: &output,
		Top:    &top,
	}

	settings.Merge(overrides)

	if settings.Output != "json" {
		t.Errorf("expected output 'json', got '%s'", settings.Output)
	}

	if settings.Top != 50 {
		t.Errorf("expected top 50, got %d", settings.Top)
	}

	// Namespace should be unchanged since it wasn't overridden
	if settings.Namespace != "default" {
		t.Errorf("expected namespace 'default', got '%s'", settings.Namespace)
	}
}

func TestMergeWithoutFlags(t *testing.T) {
	settings := &Settings{
		Output:    "yaml",
		Namespace: "kube-system",
		Top:       30,
	}

	// No overrides
	overrides := FlagOverrides{}

	settings.Merge(overrides)

	// Settings should remain unchanged
	if settings.Output != "yaml" {
		t.Errorf("expected output 'yaml', got '%s'", settings.Output)
	}

	if settings.Namespace != "kube-system" {
		t.Errorf("expected namespace 'kube-system', got '%s'", settings.Namespace)
	}

	if settings.Top != 30 {
		t.Errorf("expected top 30, got %d", settings.Top)
	}
}

func TestLoadSettingsNoFile(t *testing.T) {
	// This test will load settings when config file doesn't exist
	// It should return defaults
	settings, err := LoadSettings()

	if err != nil {
		t.Errorf("LoadSettings should not error when config file doesn't exist: %v", err)
	}

	if settings.Output != "text" {
		t.Errorf("expected default output 'text', got '%s'", settings.Output)
	}
}

func TestSaveAndLoadSettings(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()

	// Temporarily override the config path
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set HOME to temp directory for this test
	if err := os.Setenv("HOME", tempDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}

	// Create test settings
	testSettings := &Settings{
		Output:    "json",
		Namespace: "production",
		Top:       50,
		PressureThresholds: PressureThresholds{
			Low:       40.0,
			Medium:    65.0,
			High:      85.0,
			Saturated: 100.0,
		},
	}

	// Save settings
	if err := SaveSettings(testSettings); err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tempDir, ".cobrak", "settings.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config file was not created at %s", configPath)
	}

	// Load settings
	loadedSettings, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	// Verify loaded settings match saved settings
	if loadedSettings.Output != testSettings.Output {
		t.Errorf("expected output '%s', got '%s'", testSettings.Output, loadedSettings.Output)
	}

	if loadedSettings.Namespace != testSettings.Namespace {
		t.Errorf("expected namespace '%s', got '%s'", testSettings.Namespace, loadedSettings.Namespace)
	}

	if loadedSettings.Top != testSettings.Top {
		t.Errorf("expected top %d, got %d", testSettings.Top, loadedSettings.Top)
	}

	// Verify pressure thresholds were loaded
	if loadedSettings.PressureThresholds.Low != testSettings.PressureThresholds.Low {
		t.Errorf("expected Low %.1f, got %.1f", testSettings.PressureThresholds.Low, loadedSettings.PressureThresholds.Low)
	}
}

func TestSaveInvalidThresholds(t *testing.T) {
	invalidSettings := &Settings{
		Output: "text",
		PressureThresholds: PressureThresholds{
			Low:       75.0,
			Medium:    75.0, // Same as Low - invalid
			High:      90.0,
			Saturated: 100.0,
		},
	}

	// SaveSettings should fail due to invalid thresholds
	err := SaveSettings(invalidSettings)
	if err == nil {
		t.Errorf("expected SaveSettings to fail with invalid thresholds, but it succeeded")
	}
}

func TestColorConfigurationDefault(t *testing.T) {
	settings := DefaultSettings()

	if !settings.Color {
		t.Errorf("expected default Color to be true, got false")
	}
}

func TestColorConfigurationSaveAndLoad(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()

	// Set HOME to temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Create settings with color disabled
	settings := &Settings{
		Output:    "text",
		Namespace: "default",
		Top:       20,
		Color:     false,
		PressureThresholds: PressureThresholds{
			Low:       50.0,
			Medium:    75.0,
			High:      90.0,
			Saturated: 100.0,
		},
	}

	// Save settings
	if err := SaveSettings(settings); err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	// Load settings
	loadedSettings, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	// Verify color setting was loaded
	if loadedSettings.Color != false {
		t.Errorf("expected Color false, got true")
	}
}

func TestLoadTOMLWithColorSetting(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()

	// Set HOME to temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Create TOML file with color setting
	configDir := filepath.Join(tempDir, ".cobrak")
	os.MkdirAll(configDir, 0755)

	configPath := filepath.Join(configDir, "settings.toml")
	tomlContent := `output = "json"
color = false

[pressure_thresholds]
low = 50.0
medium = 75.0
high = 90.0
saturated = 100.0
`

	if err := os.WriteFile(configPath, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write test TOML file: %v", err)
	}

	// Load settings
	settings, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	// Verify color setting was parsed correctly
	if settings.Color != false {
		t.Errorf("expected Color false from TOML, got true")
	}
	if settings.Output != "json" {
		t.Errorf("expected Output 'json', got '%s'", settings.Output)
	}
}

func TestColorDefaultWhenOmittedFromTOML(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()

	// Set HOME to temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Create TOML file without color setting
	configDir := filepath.Join(tempDir, ".cobrak")
	os.MkdirAll(configDir, 0755)

	configPath := filepath.Join(configDir, "settings.toml")
	tomlContent := `output = "text"
namespace = "kube-system"

[pressure_thresholds]
low = 50.0
medium = 75.0
high = 90.0
saturated = 100.0
`

	if err := os.WriteFile(configPath, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write test TOML file: %v", err)
	}

	// Load settings - color should use default (true)
	settings, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	// Verify color setting uses default
	if !settings.Color {
		t.Errorf("expected Color true when omitted from TOML (default), got false")
	}
}
