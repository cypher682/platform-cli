package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func Load(path string) (*Config, error) {
	v := viper.New()
	resolved := resolveConfigPath(path)
	if resolved != "" {
		v.SetConfigFile(resolved)
	}
	v.SetEnvPrefix("PLATFORM")
	v.AutomaticEnv()

	if resolved != "" {
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("reading config %q: %w", resolved, err)
		}
	} else {
		if p := discoverConfig(); p != "" {
			v.SetConfigFile(p)
			if err := v.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("reading config %q: %w", p, err)
			}
		} else {
			return nil, fmt.Errorf("reading config %q: %w", path, errors.New("no such file or directory"))
		}
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if c.ComposeFile == "" {
		c.ComposeFile = "docker-compose.yml"
	}
	if len(c.Services) == 0 {
		return nil, ErrNoServices
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func resolveConfigPath(p string) string {
	if p == "" {
		p = "platform.yaml"
	}
	if env := os.Getenv("PLATFORM_CONFIG"); env != "" {
		p = env
	}
	if abs, err := filepath.Abs(p); err == nil {
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}
	if _, err := os.Stat(p); err == nil {
		abs, _ := filepath.Abs(p)
		return abs
	}
	return ""
}

func discoverConfig() string {
	cands := []string{
		"platform.yaml",
		"platform.yml",
		"platform.json",
		"platform.jsonc",
		"platform.dev.yaml",
		".platform/platform.yaml",
	}
	cwd, _ := os.Getwd()
	for _, c := range cands {
		if _, err := os.Stat(c); err == nil {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	dir := cwd
	for {
		for _, c := range cands {
			p := filepath.Join(dir, c)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
