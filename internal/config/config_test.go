package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "platform.yaml")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadValid(t *testing.T) {
	p := write(t, "project: demo\nservices:\n  - name: api\n    port: 3000\n    health: http://localhost:3000/health\n")
	c, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if c.Project != "demo" || len(c.Services) != 1 || c.Services[0].Port != 3000 {
		t.Fatalf("unexpected config: %+v", c)
	}
}

func TestLoadNoServices(t *testing.T) {
	p := write(t, "project: demo\n")
	if _, err := Load(p); !errors.Is(err, ErrNoServices) {
		t.Fatalf("want ErrNoServices, got %v", err)
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load("nope.yaml"); err == nil {
		t.Fatal("expected error")
	}
}
