package k8s

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveKubeconfig_Explicit(t *testing.T) {
	got := ResolveKubeconfig("/explicit/path")
	if got != "/explicit/path" {
		t.Errorf("expected /explicit/path, got %s", got)
	}
}

func TestResolveKubeconfig_EnvFallback(t *testing.T) {
	os.Setenv("KUBECONFIG", "/env/path")
	defer os.Unsetenv("KUBECONFIG")

	got := ResolveKubeconfig("")
	if got != "/env/path" {
		t.Errorf("expected /env/path, got %s", got)
	}
}

func TestResolveKubeconfig_HomeFallback(t *testing.T) {
	os.Unsetenv("KUBECONFIG")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home dir")
	}

	got := ResolveKubeconfig("")
	expected := filepath.Join(home, ".kube", "config")
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}
