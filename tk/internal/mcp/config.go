package mcp

import (
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/reducer"
)

type Config = reducer.ReducerConfig

func LoadConfig() (*Config, error) {
	return config_pkg.LoadConfig()
}
