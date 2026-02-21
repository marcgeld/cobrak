package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultSettings(t *testing.T) {
	settings := DefaultSettings()

	if settings.Output != "text" {
		t.Errorf("expected default output 'text', got '%s'", settings.Output)
	}

	if settings.Top != 20 {
		t.Errorf("expected default top 20, got %d", settings.Top)
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
}
