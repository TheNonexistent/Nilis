package main

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	cfg "github.com/thenonexistent/nilis/internal/config"
	"github.com/thenonexistent/nilis/internal/db"
	"github.com/thenonexistent/nilis/pkg/sharding"
	"github.com/thenonexistent/nilis/pkg/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	db        *db.Database
	shard     sharding.Shard
	shards    []sharding.Shard
	shardPool map[int]*ShardClient
	config    *cfg.Config
	store.StoreServer
}

func NewServer(config *cfg.Config, shard sharding.Shard, shards []sharding.Shard) (*Server, func() error, error) {
	if config == nil {
		return nil, nil, fmt.Errorf("null configuration provided")
	}

	if isShardEmpty(shard) {
		return nil, nil, fmt.Errorf("server shard is not initialized")
	}

	if len(shards) == 0 {
		return nil, nil, fmt.Errorf("shard list should contain at least one shard")

	}

	database, err := db.NewDatabase(config.Server.DatabaseLocation)
	if err != nil {
		log.Error().Str("module", "server").Err(err).Msg("failed to create database for store")
		return nil, nil, err
	}

	return &Server{
		db:     database,
		shard:  shard,
		shards: shards,
		config: config,
	}, database.Close, nil
}

func (s *Server) InitCluster() error {
	for _, shard := range s.shards {
		if shard.ID == s.shard.ID.ID {
			continue
		}

		cnOpts := []grpc.DialOption{}

		if s.config.Server.UseTLS {
			creds, err := credentials.NewClientTLSFromFile()
		}

		conn, err := grpc.NewClient(shard.Address)
	}
}

func (s *Server) Set(ctx context.Context, in *store.Value) (*emptypb.Empty, error) {
	err := s.db.SetKey(in.Key, in.Value)
	if err != nil {
		log.Error().Str("module", "server").Str("key", in.Key).Bytes("value", in.Value).Err(err).Msg("failed setting value in local database")
		return nil, status.Error(codes.Internal, "failed setting data in database")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) Get(ctx context.Context, in *store.Key) (*store.Value, error) {
	value, err := s.db.GetKey(in.Key)
	if err != nil {
		log.Error().Str("module", "server").Str("key", in.Key).Err(err).Msg("failed getting value from local database")
		return nil, status.Error(codes.Internal, "failed getting data from database")

	}

	if value == nil {
		return nil, status.Errorf(codes.NotFound, "key %s not found", in.Key)
	}

	return &store.Value{
		Key:   in.Key,
		Value: value,
	}, nil
}

func (s *Server) Delete(ctx context.Context, in *store.Key) (*emptypb.Empty, error) {
	err := s.db.DeleteKey(in.Key)
	if err != nil {
		log.Error().Str("module", "server").Str("key", in.Key).Err(err).Msg("failed deleting value from local database")
		return nil, status.Error(codes.Internal, "failed deleting data from database")
	}

	return &emptypb.Empty{}, nil
}

func isShardEmpty(shard sharding.Shard) bool {
	return shard.ID == 0 && shard.Address == ""
}
