//go:build integration

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/cypher682/platform-cli/internal/docker"
)

// Requires a running Docker daemon: go test -tags=integration ./tests/...
func TestDaemonReachable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := docker.CheckDaemon(ctx); err != nil {
		t.Skipf("docker unavailable: %v", err)
	}
	cli, err := docker.NewClient(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
}
