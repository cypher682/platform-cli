package scaffold

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func read(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestServiceFastAPI(t *testing.T) {
	root := t.TempDir()
	dest, err := Service(Spec{Name: "billing", Language: "fastapi", Port: 8100}, root)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"app/main.py", "Dockerfile", "docker-compose.snippet.yml", "helm/Chart.yaml", "helm/templates/deployment.yaml"} {
		if _, err := os.Stat(filepath.Join(dest, f)); err != nil {
			t.Errorf("missing %s", f)
		}
	}
	if !strings.Contains(read(t, filepath.Join(dest, "Dockerfile")), "8100") {
		t.Error("port not rendered")
	}
	// Helm syntax must survive rendering untouched.
	if !strings.Contains(read(t, filepath.Join(dest, "helm/templates/deployment.yaml")), "{{ .Values.replicaCount }}") {
		t.Error("helm template syntax was clobbered")
	}
}

func TestServiceValidation(t *testing.T) {
	root := t.TempDir()
	if _, err := Service(Spec{Name: "Bad Name", Language: "fastapi", Port: 1}, root); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("got %v", err)
	}
	if _, err := Service(Spec{Name: "ok", Language: "rust", Port: 1}, root); !errors.Is(err, ErrUnknownLang) {
		t.Fatalf("got %v", err)
	}
	if _, err := Service(Spec{Name: "ok", Language: "express", Port: 1}, root); err != nil {
		t.Fatal(err)
	}
	if _, err := Service(Spec{Name: "ok", Language: "express", Port: 1}, root); !errors.Is(err, ErrExists) {
		t.Fatalf("want ErrExists, got %v", err)
	}
}

func TestRoute(t *testing.T) {
	root := t.TempDir()
	dest, _ := Service(Spec{Name: "svc", Language: "fastapi", Port: 8000}, root)
	file, hint, err := Route(dest, "user-profile")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filepath.ToSlash(file), "app/routers/user_profile.py") || !strings.Contains(hint, "include_router") {
		t.Fatalf("%s | %s", file, hint)
	}
	if _, _, err := Route(t.TempDir(), "x"); !errors.Is(err, ErrNoService) {
		t.Fatalf("got %v", err)
	}
}

func TestGitInit(t *testing.T) {
	dest, _ := Service(Spec{Name: "svc", Language: "express", Port: 3000}, t.TempDir())
	if err := GitInit(dest, "svc"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, ".git")); err != nil {
		t.Fatal(".git missing")
	}
}
