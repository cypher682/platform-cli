package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cypher682/platform-cli/internal/scaffold"
)

var (
	scName, scLang, scDir string
	scPort                int
	scNoGit               bool
)

var scaffoldCmd = &cobra.Command{Use: "scaffold", Short: "Generate services and routes"}

// ask prompts on stdin when a flag was not provided.
func ask(r *bufio.Reader, label, def string) string {
	if def != "" {
		fmt.Printf("%s [%s]: ", label, def)
	} else {
		fmt.Printf("%s: ", label)
	}
	line, _ := r.ReadString('\n')
	if line = strings.TrimSpace(line); line != "" {
		return line
	}
	return def
}

var scaffoldServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Generate a new microservice (Dockerfile, compose snippet, Helm stub)",
	RunE: func(cmd *cobra.Command, _ []string) error {
		in := bufio.NewReader(os.Stdin)
		if scLang == "" {
			scLang = ask(in, "Language ("+strings.Join(scaffold.Languages, "/")+")", "fastapi")
		}
		if scName == "" {
			scName = ask(in, "Service name", "")
		}
		if !cmd.Flags().Changed("port") {
			def := "8000"
			if scLang == "express" {
				def = "3000"
			}
			p, err := strconv.Atoi(ask(in, "Port", def))
			if err != nil {
				return fmt.Errorf("invalid port: %w", err)
			}
			scPort = p
		}

		dest, err := scaffold.Service(scaffold.Spec{Name: scName, Language: scLang, Port: scPort}, scDir)
		if err != nil {
			return err
		}
		if !scNoGit {
			if err := scaffold.GitInit(dest, scName); err != nil {
				return fmt.Errorf("git init: %w", err)
			}
		}
		fmt.Printf("\ncreated %s\n\nnext steps:\n", dest)
		fmt.Printf("  1. add %s/docker-compose.snippet.yml to your compose file\n", dest)
		fmt.Printf("  2. add the service to platform.yaml (name, port %d, health http://localhost:%d/health)\n", scPort, scPort)
		fmt.Printf("  3. platform-cli env up\n")
		return nil
	},
}

var scaffoldRouteCmd = &cobra.Command{
	Use:   "route <name>",
	Short: "Generate a new API route in an existing service",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		file, hint, err := scaffold.Route(scDir, args[0])
		if err != nil {
			return err
		}
		fmt.Printf("created %s\n\n%s\n", file, hint)
		return nil
	},
}

func init() {
	f := scaffoldServiceCmd.Flags()
	f.StringVar(&scName, "name", "", "service name")
	f.StringVar(&scLang, "language", "", "fastapi or express")
	f.IntVar(&scPort, "port", 0, "service port")
	f.StringVar(&scDir, "dir", ".", "parent directory")
	f.BoolVar(&scNoGit, "no-git", false, "skip git init + initial commit")
	scaffoldRouteCmd.Flags().StringVar(&scDir, "dir", ".", "service directory")
	scaffoldCmd.AddCommand(scaffoldServiceCmd, scaffoldRouteCmd)
	rootCmd.AddCommand(scaffoldCmd)
}
