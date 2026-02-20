package kubeconfig_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/marcgeld/cobrak/pkg/kubeconfig"
)

func TestResolve(t *testing.T) {
	dir := t.TempDir()

	existingFile := filepath.Join(dir, "kubeconfig")
	if err := os.WriteFile(existingFile, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	otherFile := filepath.Join(dir, "other")
	if err := os.WriteFile(otherFile, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	absent := filepath.Join(dir, "does-not-exist")
	homeWithConfig := dir
	homeNoConfig := filepath.Join(dir, "nohome")

	fileExists := func(p string) bool {
		_, err := os.Stat(p)
		return err == nil
	}

	tests := []struct {
		name       string
		flagPath   string
		env        string
		homePath   string
		wantPath   string
		wantErrIs  error
	}{
		{
			name:      "flag takes priority over env and default",
			flagPath:  existingFile,
			env:       otherFile,
			homePath:  homeWithConfig,
			wantPath:  existingFile,
		},
		{
			name:      "flag ignored when file does not exist, env used",
			flagPath:  absent,
			env:       existingFile,
			homePath:  homeWithConfig,
			wantPath:  existingFile,
		},
		{
			name:      "env fallback when flag empty",
			flagPath:  "",
			env:       existingFile,
			homePath:  homeWithConfig,
			wantPath:  existingFile,
		},
		{
			name:      "env multiple paths - first non-existing then existing",
			flagPath:  "",
			env:       absent + string(os.PathListSeparator) + existingFile,
			homePath:  homeWithConfig,
			wantPath:  existingFile,
		},
		{
			name:      "env multiple paths - first existing chosen",
			flagPath:  "",
			env:       existingFile + string(os.PathListSeparator) + otherFile,
			homePath:  homeWithConfig,
			wantPath:  existingFile,
		},
		{
			name:      "default fallback when flag and env empty",
			flagPath:  "",
			env:       "",
			homePath:  homeWithConfig,
			wantPath:  filepath.Join(dir, ".kube", "config"),
		},
		{
			name:      "error when nothing found",
			flagPath:  "",
			env:       "",
			homePath:  homeNoConfig,
			wantErrIs: kubeconfig.ErrKubeconfigNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// For the default fallback test, create ~/.kube/config under the temp home.
			if tc.name == "default fallback when flag and env empty" {
				kubeDir := filepath.Join(dir, ".kube")
				if err := os.MkdirAll(kubeDir, 0o700); err != nil {
					t.Fatal(err)
				}
				kubeconfigPath := filepath.Join(kubeDir, "config")
				if err := os.WriteFile(kubeconfigPath, []byte(""), 0o600); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { os.RemoveAll(kubeDir) })
			}

			r := &kubeconfig.DefaultResolver{
				FileExists: fileExists,
				Getenv: func(key string) string {
					if key == "KUBECONFIG" {
						return tc.env
					}
					return ""
				},
				HomeDir: func() (string, error) {
					return tc.homePath, nil
				},
			}

			got, err := r.Resolve(tc.flagPath)

			if tc.wantErrIs != nil {
				if !errors.Is(err, tc.wantErrIs) {
					t.Errorf("Resolve() error = %v, want errors.Is(%v)", err, tc.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve() unexpected error: %v", err)
			}
			if got != tc.wantPath {
				t.Errorf("Resolve() = %q, want %q", got, tc.wantPath)
			}
		})
	}
}
