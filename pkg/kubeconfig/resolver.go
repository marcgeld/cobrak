package kubeconfig

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ErrKubeconfigNotFound is returned when no kubeconfig file can be located.
var ErrKubeconfigNotFound = errors.New("kubeconfig not found")

// Resolver resolves the path to a kubeconfig file.
type Resolver interface {
	Resolve(flagPath string) (string, error)
}

// DefaultResolver implements Resolver with injectable dependencies for testability.
type DefaultResolver struct {
	FileExists func(string) bool
	Getenv     func(string) string
	HomeDir    func() (string, error)
}

// NewDefaultResolver returns a DefaultResolver wired with real OS calls.
func NewDefaultResolver() *DefaultResolver {
	return &DefaultResolver{
		FileExists: func(p string) bool {
			_, err := os.Stat(p)
			return err == nil
		},
		Getenv:  os.Getenv,
		HomeDir: os.UserHomeDir,
	}
}

// Resolve returns the kubeconfig path using the following priority:
//  1. flagPath (if non-empty and the file exists)
//  2. KUBECONFIG env var (first existing path in the list)
//  3. ~/.kube/config
func (r *DefaultResolver) Resolve(flagPath string) (string, error) {
	// 1. --kubeconfig flag
	if flagPath != "" {
		if r.FileExists(flagPath) {
			return flagPath, nil
		}
	}

	// 2. KUBECONFIG environment variable (multiple paths separated by os.PathListSeparator)
	if env := r.Getenv("KUBECONFIG"); env != "" {
		for _, p := range strings.Split(env, string(os.PathListSeparator)) {
			if p != "" && r.FileExists(p) {
				return p, nil
			}
		}
	}

	// 3. ~/.kube/config
	home, err := r.HomeDir()
	if err == nil {
		defaultPath := filepath.Join(home, ".kube", "config")
		if r.FileExists(defaultPath) {
			return defaultPath, nil
		}
	}

	return "", ErrKubeconfigNotFound
}
