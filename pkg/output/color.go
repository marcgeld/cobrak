package output

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

// Global color control
var globalColorEnabled = true

func init() {
	// Initialize color support based on terminal capabilities
	SetGlobalColorEnabled(true) // Default to enabled, commands will override if needed
}

// SetGlobalColorEnabled sets the global color enabled state
func SetGlobalColorEnabled(enabled bool) {
	globalColorEnabled = enabled && isColorSupported()
	color.NoColor = !globalColorEnabled
}

// IsGlobalColorEnabled returns whether colors are globally enabled
func IsGlobalColorEnabled() bool {
	return globalColorEnabled
}

// ColorProvider handles color output based on configuration
type ColorProvider struct {
	enabled bool
}

// NewColorProvider creates a new ColorProvider
func NewColorProvider(colorEnabled bool) *ColorProvider {
	cp := &ColorProvider{
		enabled: colorEnabled && isColorSupported(),
	}
	// Also set global color state
	SetGlobalColorEnabled(colorEnabled)
	return cp
}

// isColorSupported checks if the terminal supports colors
func isColorSupported() bool {
	// Check for NO_COLOR environment variable (highest priority)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check for CLICOLOR_FORCE (force colors)
	if os.Getenv("CLICOLOR_FORCE") != "" && os.Getenv("CLICOLOR_FORCE") != "0" {
		return true
	}

	// Check if output is a terminal (using mode check)
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	// Check if stdout is a character device (terminal)
	// On Unix-like systems, terminals have mode with ModeCharDevice bit set
	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false // Not a terminal, likely piped
	}

	// Check for TERM environment variable
	term := os.Getenv("TERM")
	if term == "dumb" || term == "" {
		return false
	}

	return true
}

// Colorize applies color to text if colors are enabled
func (cp *ColorProvider) Colorize(text string, colorFunc func(string, ...interface{}) string) string {
	if cp.enabled {
		return colorFunc(text)
	}
	return text
}

// Print prints colored text if colors are enabled
func (cp *ColorProvider) Print(w io.Writer, text string, colorFunc func(string, ...interface{}) string) {
	fmt.Fprint(w, cp.Colorize(text, colorFunc))
}

// Println prints colored text with newline if colors are enabled
func (cp *ColorProvider) Println(w io.Writer, text string, colorFunc func(string, ...interface{}) string) {
	fmt.Fprintln(w, cp.Colorize(text, colorFunc))
}

// Printf prints formatted colored text if colors are enabled
func (cp *ColorProvider) Printf(w io.Writer, format string, colorFunc func(string, ...interface{}) string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	fmt.Fprint(w, cp.Colorize(text, colorFunc))
}

// IsEnabled returns whether colors are enabled
func (cp *ColorProvider) IsEnabled() bool {
	return cp.enabled
}

// Color functions that can be used with ColorProvider
// These are predefined color functions for common use cases

// Success returns green colored text
func Success(text string, args ...interface{}) string {
	return color.GreenString(text, args...)
}

// Error returns red colored text
func Error(text string, args ...interface{}) string {
	return color.RedString(text, args...)
}

// Warning returns yellow colored text
func Warning(text string, args ...interface{}) string {
	return color.YellowString(text, args...)
}

// Info returns blue colored text
func Info(text string, args ...interface{}) string {
	return color.BlueString(text, args...)
}

// Header returns cyan colored text
func Header(text string, args ...interface{}) string {
	return color.CyanString(text, args...)
}

// Bold returns bold text
func Bold(text string, args ...interface{}) string {
	return color.New(color.Bold).SprintfFunc()(text, args...)
}

// Pressure level colors
func PressureLowColor(text string) string {
	return color.GreenString(text)
}

func PressureMediumColor(text string) string {
	return color.YellowString(text)
}

func PressureHighColor(text string) string {
	return color.MagentaString(text)
}

func PressureSaturatedColor(text string) string {
	return color.RedString(text)
}

// StatusColors for different statuses
func StatusHealthy(text string) string {
	return color.GreenString(text)
}

func StatusWarning(text string) string {
	return color.YellowString(text)
}

func StatusCritical(text string) string {
	return color.RedString(text)
}

// Table colors
func TableHeader(text string) string {
	return color.CyanString(text)
}

func TableRow(text string) string {
	return text
}
