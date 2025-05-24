package sharding

import "hash/fnv"

func HashSumFromKey(key string) uint64 {
	h := fnv.New64()
	h.Write([]byte(key))
	return h.Sum64()
}

func ShardFromKey(key string, shards []Shard) Shard {
	hashSum := HashSumFromKey(key)
	return ShardFromHashSum(hashSum, shards)

}

func ShardFromHashSum(hashSum uint64, shards []Shard) Shard {
	shardID := hashSum % uint64(len(shards))
	shard, _ := FindShardById(shards, int(shardID))

	return shard
}
