package main

import (
	cfg "github.com/thenonexistent/nilis/internal/config"
	"github.com/thenonexistent/nilis/pkg/sharding"
)

func createShardsFromConfig(config *cfg.Config) ([]sharding.Shard, error) {
	shards := make([]sharding.Shard, 0, len(config.Sharding.Shards))
	for _, cfgShard := range config.Sharding.Shards {
		replicas := make([]sharding.Replica, 0, len(cfgShard.Replicas))

		for _, replicaAddress := range cfgShard.Replicas {
			replicas = append(replicas, sharding.Replica{
				Address: replicaAddress,
			})
		}

		shards = append(shards, sharding.Shard{
			ID:       cfgShard.ID,
			Address:  cfgShard.Address,
			Replicas: replicas,
		})
	}

	return shards, nil
}
