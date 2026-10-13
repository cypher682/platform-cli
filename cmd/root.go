package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cypher682/platform-cli/internal/output"
)

// cfgFile is bound to the persistent --config flag.
var cfgFile string

var globalOpts struct {
	Format  output.Format
	Quiet   bool
	NoColor bool
}

var verbose bool

var rootCmd = &cobra.Command{
	Use:          "platform-cli",
	Short:        "Developer platform CLI for local environments, health checks and logs",
	Long: `platform-cli automates common developer and operator tasks:
environment provisioning, service health checks, log tailing,
database migrations and service scaffolding.`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "platform.yaml", "path to config file")
	rootCmd.PersistentFlags().VarP(&formatFlag{&globalOpts.Format}, "format", "f", "output format (text|json)")
	rootCmd.PersistentFlags().BoolVarP(&globalOpts.Quiet, "quiet", "q", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&globalOpts.NoColor, "no-color", false, "disable colored output")

	outFmt = output.New(output.Options{Format: globalOpts.Format, Quiet: globalOpts.Quiet, NoColor: globalOpts.NoColor})
}

type formatFlag struct {
	target *output.Format
}

func (f formatFlag) String() string {
	if *f.target == "" {
		return string(output.Text)
	}
	return string(*f.target)
}
func (f formatFlag) Set(s string) error {
	switch s {
	case "json":
		*f.target = output.JSON
	default:
		*f.target = output.Text
	}
	return nil
}
func (f formatFlag) Type() string { return "string" }
