package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/cypher682/platform-cli/internal/config"
	"github.com/cypher682/platform-cli/internal/health"
)

var (
	checkTimeout  time.Duration
	watchInterval time.Duration
)

var healthCmd = &cobra.Command{Use: "health", Short: "Service health checks"}

var healthCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check all services once",
	RunE:  runHealthCheck,
}

func runHealthCheck(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}
	rs := health.CheckAll(cmd.Context(), cfg.Services, checkTimeout)
	health.PrintResults(rs)
	for _, r := range rs {
		if r.Status != health.Healthy && r.Status != health.Skipped && r.Status != health.Unknown {
			return fmt.Errorf("one or more services unhealthy")
		}
	}
	return nil
}

var healthWatchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Continuously check services; alert on status change",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()

		prev := map[string]health.Status{}
		ticker := time.NewTicker(watchInterval)
		defer ticker.Stop()

		for {
			rs := health.CheckAll(ctx, cfg.Services, checkTimeout)
			fmt.Printf("\n%s\n", time.Now().Format("15:04:05"))
			health.PrintResults(rs)

			for _, r := range rs {
				old, seen := prev[r.Service.Name]
				if seen && old != r.Status {
					fmt.Print("\a")
					color.Yellow("status change: %s %s -> %s", r.Service.Name, old, r.Status)
				}
				prev[r.Service.Name] = r.Status
			}

			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
			}
		}
	},
}

func init() {
	healthCmd.PersistentFlags().DurationVar(&checkTimeout, "timeout", 3*time.Second, "per-service check timeout")
	healthWatchCmd.Flags().DurationVarP(&watchInterval, "interval", "i", 5*time.Second, "refresh interval")
	healthCmd.AddCommand(healthCheckCmd, healthWatchCmd)
	rootCmd.AddCommand(healthCmd)
}
