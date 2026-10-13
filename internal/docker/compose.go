package docker

import (
	"context"
	"os/exec"
	"strings"
)

func detectCompose() string {
	// try docker compose first
	if err := exec.Command("docker", "compose", "version").Run(); err == nil {
		return "docker"
	}
	if err := exec.Command("docker-compose", "version").Run(); err == nil {
		return "docker-compose"
	}
	return "docker"
}

type Compose struct {
	File        string
	ProjectName string
	Profile     string
	EnvFile     string
}

func (c Compose) bin() string {
	return detectCompose()
}

func (c Compose) args(rest ...string) []string {
	bin := c.bin()
	if bin == "docker" {
		a := []string{"compose"}
		if c.File != "" {
			a = append(a, "-f", c.File)
		}
		if c.ProjectName != "" {
			a = append(a, "-p", c.ProjectName)
		}
		if c.Profile != "" {
			a = append(a, "--profile", c.Profile)
		}
		if c.EnvFile != "" {
			a = append(a, "--env-file", c.EnvFile)
		}
		return append(a, rest...)
	}
	a := []string{}
	if c.File != "" {
		a = append(a, "-f", c.File)
	}
	if c.ProjectName != "" {
		a = append(a, "-p", c.ProjectName)
	}
	if c.Profile != "" {
		a = append(a, "--profile", c.Profile)
	}
	if c.EnvFile != "" {
		a = append(a, "--env-file", c.EnvFile)
	}
	return append(a, rest...)
}

func (c Compose) Up(ctx context.Context) error {
	b := c.bin()
	if strings.HasPrefix(b, "docker") {
		return Run(ctx, "docker", c.args("up", "-d")...)
	}
	return Run(ctx, "docker-compose", c.args("up", "-d")...)
}

func (c Compose) Down(ctx context.Context, volumes bool) error {
	b := c.bin()
	a := []string{}
	if volumes {
		a = append(a, "down", "-v")
	} else {
		a = append(a, "down")
	}
	if strings.HasPrefix(b, "docker") {
		return Run(ctx, "docker", c.args(a...)...)
	}
	return Run(ctx, "docker-compose", c.args(a...)...)
}

func (c Compose) Exec(ctx context.Context, service, command string) error {
	b := c.bin()
	if strings.HasPrefix(b, "docker") {
		return Run(ctx, "docker", c.args("exec", "-T", service, "sh", "-c", command)...)
	}
	return Run(ctx, "docker-compose", c.args("exec", "-T", service, "sh", "-c", command)...)
}
