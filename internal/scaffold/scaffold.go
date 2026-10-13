// Package scaffold renders embedded templates into new services.
package scaffold

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// go:embed bakes the templates into the binary at compile time, so the
// CLI is a single self-contained file with no template directory to ship.
//
//go:embed templates
var templatesFS embed.FS

var (
	ErrInvalidName = errors.New("invalid service name (use lowercase letters, digits and dashes, starting with a letter)")
	ErrExists      = errors.New("destination already exists")
	ErrUnknownLang = errors.New("unknown language (want fastapi or express)")
	ErrNoService   = errors.New("could not detect a fastapi or express service in this directory")

	nameRe = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
)

type Spec struct {
	Name     string
	Language string // fastapi | express
	Port     int
}

// Languages lists supported templates.
var Languages = []string{"fastapi", "express"}

// ValidName reports whether s is an acceptable service/route name.
func ValidName(s string) bool { return nameRe.MatchString(s) }

// Service renders a new service into destRoot/<name>.
func Service(spec Spec, destRoot string) (string, error) {
	if !ValidName(spec.Name) {
		return "", ErrInvalidName
	}
	if spec.Language != "fastapi" && spec.Language != "express" {
		return "", ErrUnknownLang
	}
	dest := filepath.Join(destRoot, spec.Name)
	if _, err := os.Stat(dest); err == nil {
		return "", fmt.Errorf("%w: %s", ErrExists, dest)
	}

	for _, part := range []struct{ src, dst string }{
		{"templates/" + spec.Language, dest},
		{"templates/common", dest},
		{"templates/helm", filepath.Join(dest, "helm")},
	} {
		if err := renderTree(part.src, part.dst, spec); err != nil {
			return "", err
		}
	}
	return dest, nil
}

// renderTree walks an embedded directory and writes rendered files.
// "[[ ]]" delimiters keep Helm's own "{{ }}" syntax untouched.
func renderTree(srcRoot, dstRoot string, data any) error {
	return fs.WalkDir(templatesFS, srcRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel := strings.TrimSuffix(strings.TrimPrefix(p, srcRoot+"/"), ".tmpl")
		out := filepath.Join(dstRoot, filepath.FromSlash(rel))
		raw, err := templatesFS.ReadFile(p)
		if err != nil {
			return err
		}
		return renderTo(out, string(raw), data)
	})
}

func renderTo(out, tmpl string, data any) error {
	t, err := template.New(path.Base(out)).Delims("[[", "]]").Option("missingkey=error").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parsing template for %s: %w", out, err)
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Execute(f, data)
}

// Route adds a route file to an existing service and returns the path
// plus a hint describing how to register it.
func Route(serviceDir, name string) (file, hint string, err error) {
	if !ValidName(name) {
		return "", "", ErrInvalidName
	}
	data := struct{ Name, Ident string }{name, strings.ReplaceAll(name, "-", "_")}

	var tmplPath, out string
	switch {
	case fileExists(filepath.Join(serviceDir, "app", "main.py")):
		tmplPath = "templates/route/fastapi.py.tmpl"
		out = filepath.Join(serviceDir, "app", "routers", data.Ident+".py")
		hint = fmt.Sprintf("register in app/main.py:\n  from app.routers.%[1]s import router as %[1]s_router\n  app.include_router(%[1]s_router)", data.Ident)
	case fileExists(filepath.Join(serviceDir, "package.json")):
		tmplPath = "templates/route/express.ts.tmpl"
		out = filepath.Join(serviceDir, "src", "routes", name+".ts")
		hint = fmt.Sprintf("register in src/index.ts:\n  import { %[2]sRouter } from \"./routes/%[1]s\";\n  app.use(\"/%[1]s\", %[2]sRouter);", name, data.Ident)
	default:
		return "", "", ErrNoService
	}
	if fileExists(out) {
		return "", "", fmt.Errorf("%w: %s", ErrExists, out)
	}
	raw, err := templatesFS.ReadFile(tmplPath)
	if err != nil {
		return "", "", err
	}
	if err := renderTo(out, string(raw), data); err != nil {
		return "", "", err
	}
	return out, hint, nil
}

// GitInit creates a repository and an initial commit using go-git
// (pure Go: no git binary required).
func GitInit(dir, name string) error {
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		return err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	if err := wt.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		return err
	}
	_, err = wt.Commit("chore: scaffold "+name, &git.CommitOptions{
		Author: &object.Signature{Name: "platform-cli", Email: "platform-cli@localhost", When: time.Now()},
	})
	return err
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
