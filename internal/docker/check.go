package docker

import (
	"context"
	"os/exec"
)

func CheckDaemon(ctx context.Context) error {
	return exec.CommandContext(ctx, "docker", "info").Run()
}
