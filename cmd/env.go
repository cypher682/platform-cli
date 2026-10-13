package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"

	"github.com/cypher682/platform-cli/internal/config"
	"github.com/cypher682/platform-cli/internal/docker"
	"github.com/cypher682/platform-cli/internal/health"
	"github.com/cypher682/platform-cli/internal/termutil"
)

var waitTimeout time.Duration
var envForce bool
var envVolumes bool

func composeFor(cfg *config.Config) docker.Compose {
	return docker.Compose{File: cfg.ComposeFile, ProjectName: envProject, Profile: envProfile, EnvFile: envEnvFile}
}

var envCmd = &cobra.Command{Use: "env", Short: "Manage the local dev environment"}

func envUp(cmd *cobra.Command, cfg *config.Config) error {
	ctx := cmd.Context()
	if err := docker.CheckDaemon(ctx); err != nil {
		return err
	}
	if err := composeFor(cfg).Up(ctx); err != nil {
		return fmt.Errorf("docker compose up: %w", err)
	}

	sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	sp.Suffix = " waiting for services to become healthy"
	sp.Start()
	wctx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()
	rs, err := health.WaitHealthy(wctx, cfg.Services, 3*time.Second)
	sp.Stop()
	health.PrintResults(rs)
	if err != nil {
		return fmt.Errorf("not all services healthy within %s: %w", waitTimeout, err)
	}
	return nil
}

var envUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Start services and wait until they are healthy",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		return envUp(cmd, cfg)
	},
}

var envDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop and remove the environment",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		if err := docker.CheckDaemon(cmd.Context()); err != nil {
			return err
		}
		vols := envVolumes
		return composeFor(cfg).Down(cmd.Context(), vols)
	},
}

var envStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show service health",
	RunE:  runHealthCheck,
}

var envResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Down, remove volumes, then up",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		if err := docker.CheckDaemon(cmd.Context()); err != nil {
			return err
		}
		ok, err := termutil.Confirm("This will remove volumes (data). Continue?", envForce)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("aborted")
		}
		if err := composeFor(cfg).Down(cmd.Context(), true); err != nil {
			return err
		}
		return envUp(cmd, cfg)
	},
}

func init() {
	for _, c := range []*cobra.Command{envUpCmd, envResetCmd} {
		c.Flags().DurationVar(&waitTimeout, "timeout", 90*time.Second, "max time to wait for healthy services")
	}
	envResetCmd.Flags().BoolVar(&envForce, "force", false, "force without confirmation")
	envDownCmd.Flags().BoolVar(&envVolumes, "volumes", false, "also remove named volumes")
	envCmd.AddCommand(envUpCmd, envDownCmd, envStatusCmd, envResetCmd)
	rootCmd.AddCommand(envCmd)
}
