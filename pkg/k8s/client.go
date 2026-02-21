package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ResolveKubeconfig determines the kubeconfig path from explicit flag, KUBECONFIG env, or default ~/.kube/config
func ResolveKubeconfig(explicit string) string {
	if explicit != "" {
		return explicit
	}

	if env := os.Getenv("KUBECONFIG"); env != "" {
		return env
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".kube", "config")
}

// NewRestConfig builds a REST config from kubeconfig path and context
func NewRestConfig(kubeconfigPath, context string) (*rest.Config, error) {
	resolvedPath := ResolveKubeconfig(kubeconfigPath)
	if resolvedPath == "" {
		return nil, fmt.Errorf("could not resolve kubeconfig path")
	}

	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: resolvedPath},
		&clientcmd.ConfigOverrides{CurrentContext: context},
	).ClientConfig()

	if err != nil {
		return nil, fmt.Errorf("building rest config: %w", err)
	}

	return cfg, nil
}

// NewClientFromConfig builds a Kubernetes client from a REST config
func NewClientFromConfig(cfg *rest.Config) (kubernetes.Interface, error) {
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes client: %w", err)
	}

	return client, nil
}

// NewClient builds a Kubernetes client from the given kubeconfig file path.
// It returns a kubernetes.Interface so callers can substitute a fake in tests.
func NewClient(kubeconfigPath string) (kubernetes.Interface, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("building kubeconfig: %w", err)
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes client: %w", err)
	}

	return client, nil
}
