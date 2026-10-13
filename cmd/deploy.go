package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cypher682/platform-cli/internal/config"
	"github.com/cypher682/platform-cli/internal/docker"
)

var (
	depService, depNamespace string
	depContexts              []string
	depTag                   string
	depMethod                string
	depPush                  bool
	depWait                  bool
	depTimeout               string
)

var deployCmd = &cobra.Command{Use: "deploy", Short: "Deploy to Kubernetes"}

func hasBin(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

var deployLocalCmd = &cobra.Command{
	Use:   "local",
	Short: "Build, load into minikube and helm-install one service",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		var dir string
		for _, s := range cfg.Services {
			if s.Name == depService {
				if s.Compose != "" {
					dir = s.Compose
				} else {
					dir = depService
				}
			}
		}
		if dir == "" {
			return fmt.Errorf("service %q not found in config", depService)
		}
		repo := depService
		if cfg.Registry.Local != "" {
			repo = cfg.Registry.Local + "/" + depService
		}
		tag := depTag
		if tag == "" {
			tag = "dev"
		}
		image := repo + ":" + tag
		ctx := cmd.Context()

		if depMethod == "" || depMethod == "helm" {
			if !hasBin("helm") {
				return fmt.Errorf("helm not found in PATH")
			}
			if !hasBin("minikube") && !hasBin("kind") && !hasBin("kubectl") {
				// warn but proceed
			}
			steps := [][]string{
				{"docker", "build", "-t", image, dir},
			}
			if hasBin("minikube") {
				steps = append(steps, []string{"minikube", "image", "load", image})
			}
			helmArgs := []string{"upgrade", "--install", depService, filepath.Join(dir, "helm"),
				"--namespace", depNamespace, "--create-namespace",
				"--set", "image.repository=" + repo,
				"--set", "image.tag=" + tag,
			}
			if !hasBin("minikube") {
				helmArgs = append(helmArgs, "--set", "image.pullPolicy=IfNotPresent")
			} else {
				helmArgs = append(helmArgs, "--set", "image.pullPolicy=Never")
			}
			steps = append(steps, append([]string{"helm"}, helmArgs...))
			if depWait && hasBin("kubectl") {
				steps = append(steps, []string{"kubectl", "rollout", "status", "deployment/" + depService, "-n", depNamespace, "--timeout=" + depTimeout})
			}
			for _, s := range steps {
				fmt.Println("->", strings.Join(s, " "))
				if err := docker.Run(ctx, s[0], s[1:]...); err != nil {
					return fmt.Errorf("%s failed: %w", s[0], err)
				}
			}
			return nil
		}
		return fmt.Errorf("unsupported deploy method %q (supported: helm)", depMethod)
	},
}

var deployStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show deployments in one or more kubectl contexts",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctxs := depContexts
		if len(ctxs) == 0 {
			ctxs = []string{""}
		}
		for _, c := range ctxs {
			args := []string{}
			label := "current context"
			if c != "" {
				args = append(args, "--context", c)
				label = c
			}
			args = append(args, "get", "deployments", "-n", depNamespace, "-o", "wide")
			fmt.Printf("== %s ==\n", label)
			if err := docker.Run(cmd.Context(), "kubectl", args...); err != nil {
				return fmt.Errorf("kubectl (%s): %w", label, err)
			}
		}
		return nil
	},
}

func init() {
	deployCmd.PersistentFlags().StringVarP(&depNamespace, "namespace", "n", "default", "kubernetes namespace")
	deployLocalCmd.Flags().StringVar(&depService, "service", "", "service name from platform.yaml")
	_ = deployLocalCmd.MarkFlagRequired("service")
	deployLocalCmd.Flags().StringVar(&depTag, "tag", "", "image tag (default dev)")
	deployLocalCmd.Flags().StringVar(&depMethod, "method", "helm", "deploy method (helm)")
	deployLocalCmd.Flags().BoolVar(&depPush, "push", false, "push image after build")
	deployLocalCmd.Flags().BoolVar(&depWait, "wait", false, "wait for rollout")
	deployLocalCmd.Flags().StringVar(&depTimeout, "timeout", "5m", "rollout timeout")
	deployStatusCmd.Flags().StringSliceVar(&depContexts, "context", nil, "kubectl context(s)")
	deployCmd.AddCommand(deployLocalCmd, deployStatusCmd)
	rootCmd.AddCommand(deployCmd)
}
