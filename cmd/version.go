package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Set at build time:
//
//	go build -ldflags "-X github.com/cypher682/platform-cli/cmd.version=1.0.0"
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version and build info",
	Run: func(*cobra.Command, []string) {
		fmt.Printf("platform-cli %s\ncommit: %s\nbuilt:  %s\ngo:     %s %s/%s\n",
			version, commit, date, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	},
}

func init() { rootCmd.AddCommand(versionCmd) }
