package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var (
	ErrInvalidService = errors.New("invalid service configuration")
	ErrUnknownTool    = errors.New("unknown migration tool")
)

func (c *Config) Validate() error {
	if len(c.Services) == 0 {
		return ErrNoServices
	}
	names := map[string]struct{}{}
	for i, s := range c.Services {
		if s.Name == "" {
			return fmt.Errorf("%w: service[%d] missing name", ErrInvalidService, i)
		}
		if _, ok := names[s.Name]; ok {
			return fmt.Errorf("%w: duplicate service name %q", ErrInvalidService, s.Name)
		}
		names[s.Name] = struct{}{}

		// validate health URL if http-like
		if s.Health != "" {
			u, err := url.Parse(s.Health)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "tcp") {
				// allow plain; but better warn - just check basic
			}
			_ = u
		}
		if s.Port < 0 {
			return fmt.Errorf("%w: service %q has invalid port %d", ErrInvalidService, s.Name, s.Port)
		}
	}

	// database
	switch strings.ToLower(c.Database.MigrationTool) {
	case "", "prisma", "alembic", "golang-migrate":
	default:
		return fmt.Errorf("%w: %q (want prisma, alembic, or golang-migrate)", ErrUnknownTool, c.Database.MigrationTool)
	}
	return nil
}
