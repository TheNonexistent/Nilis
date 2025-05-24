package sharding

type Shard struct {
	ID       int
	Address  string
	Replicas []Replica
}

type Replica struct {
	Address string
}

func FindShardById(shards []Shard, id int) (Shard, bool) {
	for _, shard := range shards {
		if shard.ID == id {
			return shard, true
		}
	}

	return Shard{}, false
}
