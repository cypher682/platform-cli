package migrate

import (
	"errors"
	"strings"
	"testing"
)

func TestFor(t *testing.T) {
	for _, tool := range []string{"prisma", "alembic", "golang-migrate"} {
		m, err := For(tool)
		if err != nil || m.Name() != tool {
			t.Fatalf("%s: %v %v", tool, m, err)
		}
	}
	if _, err := For("flyway"); !errors.Is(err, ErrUnknownTool) {
		t.Fatalf("want ErrUnknownTool, got %v", err)
	}
}

func TestCommands(t *testing.T) {
	p, _ := For("prisma")
	if !strings.Contains(p.Migrate(), "prisma migrate deploy") {
		t.Fatal(p.Migrate())
	}
	if _, err := p.Rollback(); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("prisma rollback should be unsupported, got %v", err)
	}
	a, _ := For("alembic")
	if c, err := a.Rollback(); err != nil || c != "alembic downgrade -1" {
		t.Fatalf("%q %v", c, err)
	}
}
