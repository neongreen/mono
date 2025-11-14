package mcp

import (
	config_pkg "github.com/neongreen/mono/tk/internal/config"
)

type Config = config_pkg.Config

func LoadConfig() (*Config, error) {
	return config_pkg.LoadConfig()
}
