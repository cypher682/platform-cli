package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cypher682/platform-cli/internal/config"
	"github.com/cypher682/platform-cli/internal/migrate"
	"github.com/cypher682/platform-cli/internal/termutil"
)

var (
	dbForce   bool
	dbService string
)

var dbCmd = &cobra.Command{Use: "db", Short: "Database migrations"}

func runDB(cmd *cobra.Command, pick func(migrate.Migrator) (string, error), showHint bool, needsConfirm bool) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}
	m, err := migrate.For(cfg.Database.MigrationTool)
	if err != nil {
		return err
	}
	command, err := pick(m)
	if err != nil {
		return err
	}
	svc := dbService
	if svc == "" {
		svc = cfg.Database.Service
	}
	if svc == "" {
		svc = cfg.Database.Container
	}
	if svc == "" {
		return errors.New("config: set database.service (or database.container)")
	}
	if needsConfirm {
		ok, err := termutil.Confirm(fmt.Sprintf("Run rollback via %s in %s?", m.Name(), svc), dbForce)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("aborted")
		}
	}
	fmt.Printf("-> [%s] %s\n", svc, command)
	if err := composeFor(cfg).Exec(cmd.Context(), svc, command); err != nil {
		if showHint {
			fmt.Println("hint:", m.FailureHint())
		}
		return fmt.Errorf("%s failed: %w", m.Name(), err)
	}
	return nil
}

func dbSub(use, short string, hint, confirm bool, pick func(migrate.Migrator) (string, error)) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE:  func(cmd *cobra.Command, _ []string) error { return runDB(cmd, pick, hint, confirm) },
	}
}

func init() {
	dbCmd.AddCommand(
		dbSub("migrate", "Run pending migrations", true, false, func(m migrate.Migrator) (string, error) { return m.Migrate(), nil }),
		dbSub("rollback", "Roll back the last migration", false, true, func(m migrate.Migrator) (string, error) { return m.Rollback() }),
		dbSub("seed", "Run the seed script", false, false, func(m migrate.Migrator) (string, error) { return m.Seed(), nil }),
		dbSub("status", "Show migration status", false, false, func(m migrate.Migrator) (string, error) { return m.Status(), nil }),
	)
	dbCmd.PersistentFlags().BoolVar(&dbForce, "force", false, "force without confirmation")
	dbCmd.PersistentFlags().StringVar(&dbService, "service", "", "override database service/container")
	rootCmd.AddCommand(dbCmd)
}
