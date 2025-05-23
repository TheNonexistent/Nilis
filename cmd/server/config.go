package main

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var defaults = map[string]any{
	"server.listen_port":       6226,
	"server.bind_address":      "0.0.0.0",
	"server.database_location": "/opt/nilis/local.db",
	"server.use_tls":           false,

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
		UseTLS           bool   `mapstructure:"use_tls"`
		TLSCert          string `mapstructure:"tls_cert"`
		TLSKey           string `mapstructure:"tls_key"`
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

func validateConfig(config *Config) error {
	if config.Server.ListenPort <= 0 || config.Server.ListenPort > 65535 {
		return fmt.Errorf("invalid listen port: %d", config.Server.ListenPort)
	}
	if config.Server.BindAddress == "" {
		return errors.New("bind address cannot be empty")
	}
	if config.Server.DatabaseLocation == "" {
		return errors.New("database location cannot be empty")
	}

	if config.Server.UseTLS && config.Server.TLSCert == "" {
		return errors.New("tls certificate location cannot be empty when using tls mode")
	}

	if config.Server.UseTLS && config.Server.TLSKey == "" {
		return errors.New("tls key location cannot be empty when using tls mode")
	}

	if config.Logging.Level == "" {
		return errors.New("logging level cannot be empty")
	}
	if config.Logging.File == "" {
		return errors.New("logging file path cannot be empty")
	}

	if config.Sharding.Enabled {
		if config.Sharding.ShardID < 0 {
			return fmt.Errorf("shard_id must be non-negative")
		}

		shardIDs := make(map[int]struct{})
		shardAddresses := make(map[string]struct{})
		replicaAddresses := make(map[string]struct{})

		for _, shard := range config.Sharding.Shards {
			if shard.ID < 0 {
				return fmt.Errorf("shard id must be non-negative: %d", shard.ID)
			}
			if shard.Address == "" {
				return fmt.Errorf("shard address cannot be empty for shard id: %d", shard.ID)
			}
			if !isValidAddressFormat(shard.Address) {
				return fmt.Errorf("shard address does not follow ip:port format for shard id: %d", shard.ID)
			}

			if _, exists := shardIDs[shard.ID]; exists {
				return fmt.Errorf("duplicate shard id found: %d", shard.ID)
			}
			shardIDs[shard.ID] = struct{}{}

			if _, exists := shardAddresses[shard.Address]; exists {
				return fmt.Errorf("duplicate shard address found: %s", shard.Address)
			}
			shardAddresses[shard.Address] = struct{}{}

			for _, replica := range shard.Replicas {
				if replica == "" {
					return fmt.Errorf("replica address cannot be empty for shard id: %d", shard.ID)
				}
				if !isValidAddressFormat(replica) {
					return fmt.Errorf("replica address does not follow ip:port format for shard id: %d", shard.ID)
				}

				if _, exists := replicaAddresses[replica]; exists {
					return fmt.Errorf("duplicate replica address found in shard id %d: %s", shard.ID, replica)
				}
				replicaAddresses[replica] = struct{}{}
			}
		}

		numShards := len(config.Sharding.Shards)
		if numShards == 0 || !isPowerOfTwo(numShards) {
			return fmt.Errorf("number of shards must be a power of 2, got: %d", numShards)
		}
	}

	return nil
}

func isPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}

func isValidAddressFormat(address string) bool {
	addressRegexp := regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?):\d{1,5}\b`)

	return addressRegexp.MatchString(address)
}
