package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ResolveKubeconfig returns the effective kubeconfig path.
// Priority: explicit path > KUBECONFIG env > ~/.kube/config
func ResolveKubeconfig(explicitPath string) string {
	if explicitPath != "" {
		return explicitPath
	}
	if env := os.Getenv("KUBECONFIG"); env != "" {
		return env
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".kube", "config")
	}
	return ""
}

// NewRestConfig builds a *rest.Config from the given kubeconfig path and context.
func NewRestConfig(kubeconfigPath, kubeContext string) (*rest.Config, error) {
	resolved := ResolveKubeconfig(kubeconfigPath)
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: resolved}
	configOverrides := &clientcmd.ConfigOverrides{}
	if kubeContext != "" {
		configOverrides.CurrentContext = kubeContext
	}
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig from %q: %w", resolved, err)
	}
	return cfg, nil
}

// NewClientFromConfig builds a kubernetes.Interface from a *rest.Config.
func NewClientFromConfig(cfg *rest.Config) (kubernetes.Interface, error) {
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes client: %w", err)
	}
	return client, nil
}

// NewClient builds a kubernetes.Interface from kubeconfig path and context.
func NewClient(kubeconfigPath, kubeContext string) (kubernetes.Interface, error) {
	cfg, err := NewRestConfig(kubeconfigPath, kubeContext)
	if err != nil {
		return nil, err
	}
	return NewClientFromConfig(cfg)
}
