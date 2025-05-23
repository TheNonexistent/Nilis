package main

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var defaults = map[string]any{
	"server.listen_port":       6226,
	"server.bind_address":      "0.0.0.0",
	"server.database_location": "/opt/nilis/local.db",

	"sharding.enabled":  false,
	"sharding.shard_id": 0,
	"sharding.replica":  false,
	"sharding.shards":   []map[string]any{},

	"logging.level": "info",
	"logging.file":  "/var/log/nilis.log",
}

type Config struct {
	Server struct {
		ListenPort       int    `mapstructure:"listen_port"`
		BindAddress      string `mapstructure:"bind_address"`
		DatabaseLocation string `mapstructure:"database_location"`
	} `mapstructure:"server"`

	Sharding struct {
		Enabled bool `mapstructure:"enabled"`
		ShardID int  `mapstructure:"shard_id"`
		Replica bool `mapstructure:"replica"`
		Shards  []struct {
			ID       int      `mapstructure:"id"`
			Address  string   `mapstructure:"address"`
			Replicas []string `mapstructure:"replicas"`
		} `mapstructure:"shards"`
	} `mapstructure:"sharding"`

	Logging struct {
		Level string `mapstructure:"level"`
		File  string `mapstructure:"file"`
	} `mapstructure:"logging"`
}

func loadConfig(config *Config) error {
	v := viper.New()
	v.SetConfigName("nilis")
	v.SetConfigType("yaml")
	v.AddConfigPath("/etc/nilis")
	v.AddConfigPath(".")

	for key, value := range defaults {
		v.SetDefault(key, value)
	}

	if err := v.ReadInConfig(); err != nil {
		log.Warn().Err(err).Str("module", "configuration").Msg("could not read config file, using defaults...")
	}

	if err := v.Unmarshal(config); err != nil {
		return fmt.Errorf("error unmarshaling config file: %w", err)
	}

	return nil
}
