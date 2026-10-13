package config

import "net/url"

type HealthCheck struct {
	Type         string            `mapstructure:"type"`
	URL          string            `mapstructure:"url"`
	Path         string            `mapstructure:"path"`
	Method       string            `mapstructure:"method"`
	Headers      map[string]string `mapstructure:"headers"`
	Port         int               `mapstructure:"port"`
	Addr         string            `mapstructure:"addr"`
	ExpectStatus int               `mapstructure:"expect_status"`
	BodyContains string            `mapstructure:"body_contains"`
}

type Service struct {
	Name        string       `mapstructure:"name"`
	Port        int          `mapstructure:"port"`
	Health      string       `mapstructure:"health"`
	Compose     string       `mapstructure:"compose"`
	HealthCheck *HealthCheck `mapstructure:"health_check"`
}

type Database struct {
	Type          string `mapstructure:"type"`
	MigrationTool string `mapstructure:"migration_tool"`
	Container     string `mapstructure:"container"`
	Service       string `mapstructure:"service"`
}

type Registry struct {
	Local  string `mapstructure:"local"`
	Remote string `mapstructure:"remote"`
}

type Config struct {
	Project     string    `mapstructure:"project"`
	ComposeFile string    `mapstructure:"compose_file"`
	Services    []Service `mapstructure:"services"`
	Database    Database  `mapstructure:"database"`
	Registry    Registry  `mapstructure:"registry"`
}

// ensure backward compat
func (s *Service) Normalize() {
	if s.HealthCheck == nil && s.Health != "" {
		u, _ := url.Parse(s.Health)
		typ := "http"
		if u != nil && u.Scheme == "tcp" {
			typ = "tcp"
		}
		s.HealthCheck = &HealthCheck{Type: typ, URL: s.Health}
	}
}
