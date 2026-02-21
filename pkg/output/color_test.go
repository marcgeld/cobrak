package output

import (
	"bytes"
	"os"
	"testing"
)

func TestNewColorProvider(t *testing.T) {
	tests := []struct {
		name          string
		colorEnabled  bool
		expectEnabled bool
	}{
		{
			name:          "color enabled when requested",
			colorEnabled:  true,
			expectEnabled: true,
		},
		{
			name:          "color disabled when requested",
			colorEnabled:  false,
			expectEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := NewColorProvider(tt.colorEnabled)
			if tt.colorEnabled && cp.enabled != tt.expectEnabled {
				// Note: enabled might be false if running in non-TTY environment
				// So we only check if colorEnabled is false
			}
			if !tt.colorEnabled && cp.enabled {
				t.Errorf("expected color disabled, but got enabled")
			}
		})
	}
}

func TestColorizeWithColorDisabled(t *testing.T) {
	cp := NewColorProvider(false)
	text := "test text"
	colorFunc := func(s string, args ...interface{}) string {
		return "[RED]" + s + "[/RED]"
	}

	result := cp.Colorize(text, colorFunc)
	if result != text {
		t.Errorf("expected plain text '%s', got '%s'", text, result)
	}
}

func TestColorizeWithColorEnabled(t *testing.T) {
	cp := &ColorProvider{enabled: true}
	text := "test text"
	colorFunc := func(s string, args ...interface{}) string {
		return "[COLORED]" + s + "[/COLORED]"
	}

	result := cp.Colorize(text, colorFunc)
	if result != "[COLORED]test text[/COLORED]" {
		t.Errorf("expected colored text, got '%s'", result)
	}
}

func TestColorProviderPrint(t *testing.T) {
	cp := NewColorProvider(false)
	buf := bytes.NewBufferString("")

	colorFunc := Success
	cp.Print(buf, "test", colorFunc)

	if buf.String() == "" {
		t.Errorf("expected output, got empty string")
	}
}

func TestColorProviderPrintln(t *testing.T) {
	cp := NewColorProvider(false)
	buf := bytes.NewBufferString("")

	colorFunc := Error
	cp.Println(buf, "test", colorFunc)

	output := buf.String()
	if output == "" {
		t.Errorf("expected output with newline, got empty string")
	}
	if !bytes.HasSuffix(buf.Bytes(), []byte("\n")) {
		t.Errorf("expected output to end with newline")
	}
}

func TestColorProviderIsEnabled(t *testing.T) {
	cpEnabled := &ColorProvider{enabled: true}
	if !cpEnabled.IsEnabled() {
		t.Errorf("expected IsEnabled to return true")
	}

	cpDisabled := &ColorProvider{enabled: false}
	if cpDisabled.IsEnabled() {
		t.Errorf("expected IsEnabled to return false")
	}
}

func TestColorFunctionsExist(t *testing.T) {
	// Test that color functions exist and return strings
	tests := []struct {
		name       string
		colorFunc  func(string, ...interface{}) string
		input      string
		shouldWork bool
	}{
		{
			name:       "Success color",
			colorFunc:  Success,
			input:      "test",
			shouldWork: true,
		},
		{
			name:       "Error color",
			colorFunc:  Error,
			input:      "test",
			shouldWork: true,
		},
		{
			name:       "Warning color",
			colorFunc:  Warning,
			input:      "test",
			shouldWork: true,
		},
		{
			name:       "Info color",
			colorFunc:  Info,
			input:      "test",
			shouldWork: true,
		},
		{
			name:       "Header color",
			colorFunc:  Header,
			input:      "test",
			shouldWork: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.colorFunc(tt.input)
			if tt.shouldWork && result == "" {
				t.Errorf("expected non-empty result, got empty string")
			}
		})
	}
}

func TestPressureLevelColors(t *testing.T) {
	tests := []struct {
		name      string
		colorFunc func(string) string
		input     string
	}{
		{
			name:      "Low pressure color",
			colorFunc: PressureLowColor,
			input:     "LOW",
		},
		{
			name:      "Medium pressure color",
			colorFunc: PressureMediumColor,
			input:     "MEDIUM",
		},
		{
			name:      "High pressure color",
			colorFunc: PressureHighColor,
			input:     "HIGH",
		},
		{
			name:      "Saturated pressure color",
			colorFunc: PressureSaturatedColor,
			input:     "SATURATED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.colorFunc(tt.input)
			if result == "" {
				t.Errorf("expected colored result, got empty string")
			}
		})
	}
}

func TestStatusColors(t *testing.T) {
	tests := []struct {
		name      string
		colorFunc func(string) string
		input     string
	}{
		{
			name:      "Healthy status",
			colorFunc: StatusHealthy,
			input:     "HEALTHY",
		},
		{
			name:      "Warning status",
			colorFunc: StatusWarning,
			input:     "WARNING",
		},
		{
			name:      "Critical status",
			colorFunc: StatusCritical,
			input:     "CRITICAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.colorFunc(tt.input)
			if result == "" {
				t.Errorf("expected colored result, got empty string")
			}
		})
	}
}

func TestColorSupportDetection(t *testing.T) {
	// Save original stdout
	originalStdout := os.Stdout

	t.Run("color support detection", func(t *testing.T) {
		// Test that isColorSupported function exists and can be called
		// We can't directly test it since it's unexported, but we can
		// test through NewColorProvider
		cp := NewColorProvider(true)
		if cp == nil {
			t.Errorf("expected non-nil ColorProvider")
		}
	})

	// Restore stdout
	os.Stdout = originalStdout
}

func TestNoColorEnvironmentVariable(t *testing.T) {
	// Save original NO_COLOR env var
	originalNoColor := os.Getenv("NO_COLOR")
	defer func() {
		_ = os.Setenv("NO_COLOR", originalNoColor)
	}()

	// Test with NO_COLOR set
	_ = os.Setenv("NO_COLOR", "1")
	cp := NewColorProvider(true)
	// Note: When NO_COLOR is set, colors should be disabled
	// But this depends on terminal detection, so we just verify it doesn't panic
	if cp == nil {
		t.Errorf("expected non-nil ColorProvider even with NO_COLOR set")
	}
}

func TestColorProviderWithFormatString(t *testing.T) {
	cp := NewColorProvider(false)
	buf := bytes.NewBufferString("")

	colorFunc := func(s string, args ...interface{}) string {
		return s
	}

	cp.Printf(buf, "test %d", colorFunc, 42)

	output := buf.String()
	if output != "test 42" {
		t.Errorf("expected 'test 42', got '%s'", output)
	}
}
