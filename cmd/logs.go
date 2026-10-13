package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/cypher682/platform-cli/internal/config"
	"github.com/cypher682/platform-cli/internal/docker"
	"github.com/cypher682/platform-cli/internal/logs"
)

var (
	logSince  time.Duration
	logFilter string
	logTail   string
	logFollow bool
	logTS     bool
)

var logsCmd = &cobra.Command{Use: "logs", Short: "Tail and search service logs"}

func serviceNames(cfg *config.Config, args []string) []string {
	if len(args) > 0 {
		return args
	}
	names := make([]string, len(cfg.Services))
	for i, s := range cfg.Services {
		names[i] = s.Name
	}
	return names
}

func newPrinter(svcs []string) func(logs.Line) {
	palette := []color.Attribute{color.FgCyan, color.FgMagenta, color.FgYellow, color.FgGreen, color.FgBlue, color.FgRed}
	colors := map[string]*color.Color{}
	width := 0
	for i, s := range svcs {
		colors[s] = color.New(palette[i%len(palette)])
		if len(s) > width {
			width = len(s)
		}
	}
	return func(l logs.Line) {
		fmt.Printf("%s | %s\n", colors[l.Service].Sprintf("%-*s", width, l.Service), l.Text)
	}
}

func consume(cmd *cobra.Command, svcs []string, o logs.Options, match func(string) bool, onMatch func()) error {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer stop()

	cli, err := docker.NewClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()

	out := make(chan logs.Line)
	done := make(chan error, 1)
	go func() {
		done <- logs.FanIn(ctx, cli, svcs, o, out)
		close(out)
	}()

	emit := newPrinter(svcs)
	for l := range out {
		if match(l.Text) {
			emit(l)
			if onMatch != nil {
				onMatch()
			}
		}
	}
	return <-done
}

var logsTailCmd = &cobra.Command{
	Use:   "tail [service...]",
	Short: "Tail logs from one or more services (default: all in config)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		match, err := logs.NewMatcher(logFilter)
		if err != nil {
			return err
		}
		o := logs.Options{Since: logSince, Follow: logFollow, Tail: logTail}
		return consume(cmd, serviceNames(cfg, args), o, match, nil)
	},
}

var logsSearchCmd = &cobra.Command{
	Use:   "search <pattern> [service...]",
	Short: "Search recent logs with a regex",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		match, err := logs.NewMatcher(args[0])
		if err != nil {
			return err
		}
		o := logs.Options{Since: logSince, Follow: false, Tail: "all"}
		n := 0
		if err := consume(cmd, serviceNames(cfg, args[1:]), o, match, func() { n++ }); err != nil {
			return err
		}
		fmt.Printf("\n%d matching line(s)\n", n)
		return nil
	},
}

func init() {
	logsTailCmd.Flags().DurationVar(&logSince, "since", 0, "only logs newer than this")
	logsTailCmd.Flags().StringVar(&logFilter, "filter", "", "only lines matching this regex")
	logsTailCmd.Flags().StringVar(&logTail, "tail", "50", "lines to show from end (or \"all\")")
	logsTailCmd.Flags().BoolVar(&logFollow, "follow", true, "follow log output")
	logsTailCmd.Flags().BoolVar(&logTS, "timestamps", false, "show timestamps")
	logsSearchCmd.Flags().DurationVar(&logSince, "since", time.Hour, "search logs newer than this")
	logsSearchCmd.Flags().BoolVar(&logTS, "timestamps", false, "show timestamps")
	logsCmd.AddCommand(logsTailCmd, logsSearchCmd)
	rootCmd.AddCommand(logsCmd)
}
