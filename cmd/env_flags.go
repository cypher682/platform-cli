package cmd

var (
	envProject string
	envProfile string
	envEnvFile string
)

func init() {
	envUpCmd.Flags().StringVar(&envProject, "project-name", "", "docker compose project name")
	envUpCmd.Flags().StringVar(&envProfile, "profile", "", "docker compose profile")
	envUpCmd.Flags().StringVar(&envEnvFile, "env-file", "", "docker compose env file")
	envDownCmd.Flags().StringVar(&envProject, "project-name", "", "docker compose project name")
	envDownCmd.Flags().StringVar(&envProfile, "profile", "", "docker compose profile")
	envDownCmd.Flags().StringVar(&envEnvFile, "env-file", "", "docker compose env file")
	envResetCmd.Flags().StringVar(&envProject, "project-name", "", "docker compose project name")
	envResetCmd.Flags().StringVar(&envProfile, "profile", "", "docker compose profile")
	envResetCmd.Flags().StringVar(&envEnvFile, "env-file", "", "docker compose env file")
}
