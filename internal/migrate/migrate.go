// Package migrate abstracts migration tools behind one interface, so the
// db commands never know whether they are driving Prisma or Alembic.
package migrate

import (
	"errors"
	"fmt"
)

var (
	ErrUnknownTool = errors.New("unknown migration tool")
	ErrUnsupported = errors.New("operation not supported by this tool")
)

// Migrator returns the shell command to run inside the service container.
type Migrator interface {
	Name() string
	Migrate() string
	Rollback() (string, error)
	Seed() string
	Status() string
	FailureHint() string
}

// For picks the backend named in platform.yaml.
func For(tool string) (Migrator, error) {
	switch tool {
	case "prisma":
		return prisma{}, nil
	case "alembic":
		return alembic{}, nil
	case "golang-migrate":
		return golangMigrate{}, nil
	default:
		return nil, fmt.Errorf("%w: %q (want prisma, alembic or golang-migrate)", ErrUnknownTool, tool)
	}
}

type prisma struct{}

func (prisma) Name() string    { return "prisma" }
func (prisma) Migrate() string { return "npx prisma migrate deploy" }
func (prisma) Rollback() (string, error) {
	// Prisma has no down-migrations; be explicit rather than pretend.
	return "", fmt.Errorf("%w: prisma has no rollback; mark the migration with `prisma migrate resolve --rolled-back <name>` and restore from backup", ErrUnsupported)
}
func (prisma) Seed() string   { return "npx prisma db seed" }
func (prisma) Status() string { return "npx prisma migrate status" }
func (prisma) FailureHint() string {
	return "inspect with `platform-cli db status`; prisma has no automatic rollback"
}

type alembic struct{}

func (alembic) Name() string              { return "alembic" }
func (alembic) Migrate() string           { return "alembic upgrade head" }
func (alembic) Rollback() (string, error) { return "alembic downgrade -1", nil }
func (alembic) Seed() string              { return "python scripts/seed.py" }
func (alembic) Status() string            { return "alembic current" }
func (alembic) FailureHint() string       { return "revert with `platform-cli db rollback`" }

type golangMigrate struct{}

func (golangMigrate) Name() string { return "golang-migrate" }
func (golangMigrate) Migrate() string {
	return `migrate -path migrations -database "$DATABASE_URL" up`
}
func (golangMigrate) Rollback() (string, error) {
	return `migrate -path migrations -database "$DATABASE_URL" down 1`, nil
}
func (golangMigrate) Seed() string { return "sh scripts/seed.sh" }
func (golangMigrate) Status() string {
	return `migrate -path migrations -database "$DATABASE_URL" version`
}
func (golangMigrate) FailureHint() string {
	return "revert with `platform-cli db rollback`; fix a dirty state with `migrate force <version>`"
}
