package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/marcgeld/cobrak/pkg/config"
)

func TestResolveConfigPath_Default(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Ensure COBRAK_CONFIG is not set
	originalEnv := os.Getenv("COBRAK_CONFIG")
	defer os.Setenv("COBRAK_CONFIG", originalEnv)
	os.Unsetenv("COBRAK_CONFIG")

	got, err := config.ResolveConfigPath("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join(tempDir, ".cobrak", "settings.toml")
	if got != want {
		t.Errorf("ResolveConfigPath(\"\") = %q, want %q", got, want)
	}
}

func TestResolveConfigPath_EnvOverride(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	originalEnv := os.Getenv("COBRAK_CONFIG")
	defer os.Setenv("COBRAK_CONFIG", originalEnv)
	os.Setenv("COBRAK_CONFIG", "custom.toml")

	got, err := config.ResolveConfigPath("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join(tempDir, ".cobrak", "custom.toml")
	if got != want {
		t.Errorf("ResolveConfigPath(\"\") with COBRAK_CONFIG=custom.toml = %q, want %q", got, want)
	}
}

func TestResolveConfigPath_FlagOverride(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	originalEnv := os.Getenv("COBRAK_CONFIG")
	defer os.Setenv("COBRAK_CONFIG", originalEnv)
	os.Unsetenv("COBRAK_CONFIG")

	got, err := config.ResolveConfigPath("work.toml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join(tempDir, ".cobrak", "work.toml")
	if got != want {
		t.Errorf("ResolveConfigPath(\"work.toml\") = %q, want %q", got, want)
	}
}

func TestResolveConfigPath_FlagTakesPriorityOverEnv(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	originalEnv := os.Getenv("COBRAK_CONFIG")
	defer os.Setenv("COBRAK_CONFIG", originalEnv)
	os.Setenv("COBRAK_CONFIG", "env.toml")

	got, err := config.ResolveConfigPath("flag.toml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join(tempDir, ".cobrak", "flag.toml")
	if got != want {
		t.Errorf("flag should override env: got %q, want %q", got, want)
	}
}

func TestResolveConfigPath_RejectsAbsolutePath(t *testing.T) {
	_, err := config.ResolveConfigPath("/etc/cobrak/settings.toml")
	if !errors.Is(err, config.ErrAbsolutePath) {
		t.Errorf("expected ErrAbsolutePath, got %v", err)
	}
}

func TestResolveConfigPath_EnvRejectsAbsolutePath(t *testing.T) {
	originalEnv := os.Getenv("COBRAK_CONFIG")
	defer os.Setenv("COBRAK_CONFIG", originalEnv)
	os.Setenv("COBRAK_CONFIG", "/etc/cobrak/settings.toml")

	_, err := config.ResolveConfigPath("")
	if !errors.Is(err, config.ErrAbsolutePath) {
		t.Errorf("expected ErrAbsolutePath from env, got %v", err)
	}
}

func TestResolveConfigPath_RejectsTraversal(t *testing.T) {
	tests := []string{
		"../escape.toml",
		"../../etc/passwd",
		"sub/../../escape.toml",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			_, err := config.ResolveConfigPath(path)
			if !errors.Is(err, config.ErrPathTraversal) {
				t.Errorf("expected ErrPathTraversal for %q, got %v", path, err)
			}
		})
	}
}

func TestResolveConfigPath_EnvRejectsTraversal(t *testing.T) {
	originalEnv := os.Getenv("COBRAK_CONFIG")
	defer os.Setenv("COBRAK_CONFIG", originalEnv)
	os.Setenv("COBRAK_CONFIG", "../escape.toml")

	_, err := config.ResolveConfigPath("")
	if !errors.Is(err, config.ErrPathTraversal) {
		t.Errorf("expected ErrPathTraversal from env traversal, got %v", err)
	}
}

func TestResolveConfigPath_SubdirectoryAllowed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("path separator difference on Windows")
	}

	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	originalEnv := os.Getenv("COBRAK_CONFIG")
	defer os.Setenv("COBRAK_CONFIG", originalEnv)
	os.Unsetenv("COBRAK_CONFIG")

	got, err := config.ResolveConfigPath("sub/config.toml")
	if err != nil {
		t.Fatalf("unexpected error for subdirectory path: %v", err)
	}

	want := filepath.Join(tempDir, ".cobrak", "sub", "config.toml")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSaveSettingsAt_FilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permission check not applicable on Windows")
	}

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.toml")

	settings := config.DefaultSettings()
	if err := config.SaveSettingsAt(configPath, settings); err != nil {
		t.Fatalf("SaveSettingsAt failed: %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	// Expect file mode 0600 (owner read/write only)
	got := info.Mode().Perm()
	want := os.FileMode(0600)
	if got != want {
		t.Errorf("file permissions = %04o, want %04o", got, want)
	}
}

func TestSaveSettingsAt_DirectoryPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permission check not applicable on Windows")
	}

	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".cobrak")
	configPath := filepath.Join(configDir, "settings.toml")

	settings := config.DefaultSettings()
	if err := config.SaveSettingsAt(configPath, settings); err != nil {
		t.Fatalf("SaveSettingsAt failed: %v", err)
	}

	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	// Expect directory mode 0700 (owner read/write/execute only)
	got := info.Mode().Perm()
	want := os.FileMode(0700)
	if got != want {
		t.Errorf("directory permissions = %04o, want %04o", got, want)
	}
}

func TestLoadSettingsAt_RoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.toml")

	original := &config.Settings{
		Output:    "json",
		Namespace: "test-ns",
		Top:       42,
		Color:     false,
		PressureThresholds: config.PressureThresholds{
			Low:       30.0,
			Medium:    55.0,
			High:      80.0,
			Saturated: 95.0,
		},
	}

	if err := config.SaveSettingsAt(configPath, original); err != nil {
		t.Fatalf("SaveSettingsAt failed: %v", err)
	}

	loaded, err := config.LoadSettingsAt(configPath)
	if err != nil {
		t.Fatalf("LoadSettingsAt failed: %v", err)
	}

	if loaded.Output != original.Output {
		t.Errorf("Output: got %q, want %q", loaded.Output, original.Output)
	}
	if loaded.Namespace != original.Namespace {
		t.Errorf("Namespace: got %q, want %q", loaded.Namespace, original.Namespace)
	}
	if loaded.Top != original.Top {
		t.Errorf("Top: got %d, want %d", loaded.Top, original.Top)
	}
	if loaded.Color != original.Color {
		t.Errorf("Color: got %v, want %v", loaded.Color, original.Color)
	}
}
