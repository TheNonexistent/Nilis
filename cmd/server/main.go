package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	cfg "github.com/thenonexistent/nilis/internal/config"
	"github.com/thenonexistent/nilis/pkg/sharding"
	"github.com/thenonexistent/nilis/pkg/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var config cfg.Config

func main() {

	if err := cfg.LoadConfig(&config); err != nil {
		log.Fatal().Str("module", "main").Err(err).Msg("error in initial configuration")
	}

	if err := cfg.ValidateConfig(&config); err != nil {
		log.Fatal().Str("module", "main").Err(err).Msg("invalid configuration")
	}

	initLogger(config.Logging.Level)

	var shards []sharding.Shard
	var serverShard sharding.Shard

	if config.Sharding.Enabled {
		var err error
		var ok bool

		shards, err = createShardsFromConfig(&config)
		if err != nil {
			log.Fatal().Err(err).Msg("failed creating shards based on provided configuraion")
		}

		serverShard, ok = sharding.FindShardById(shards, config.Sharding.ShardID)
		if !ok {
			log.Fatal().Int("shard_id", config.Sharding.ShardID).Msg("provided shard id is not present withing sharding configuration")
		}
	} else {
		serverShard = sharding.Shard{
			ID:       0,
			Address:  fmt.Sprintf("%s:%d", config.Server.BindAddress, config.Server.ListenPort),
			Replicas: []sharding.Replica{},
		}
		shards = []sharding.Shard{serverShard}
	}

	storeServer, cancelFunc, err := NewServer(&config, serverShard, shards)
	if err != nil {
		log.Fatal().Str("module", "main").Err(err).Msg("failed to create store server")
	}
	defer cancelFunc()

	log.Info().Int("shard_id", serverShard.ID).Msg("initialized server")

	errChan := make(chan error, 2)
	sigChan := make(chan os.Signal, 1)

	go func() {
		if err := startGRPCServer(storeServer); err != nil {
			errChan <- fmt.Errorf("gRPC server failed: %w", err)
		}
	}()

	select {
	case sig := <-sigChan:
		log.Info().Str("signal", sig.String()).Msg("received interrupt, exiting...")
		os.Exit(1)

	case err := <-errChan:
		log.Fatal().Str("module", "main").Err(err).Msg("store service failed")
	}
}

func startGRPCServer(storeServer *Server) error {
	ListenAddr := fmt.Sprintf("%s:%d", config.Server.BindAddress, config.Server.ListenPort)
	lis, err := net.Listen("tcp", ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", ListenAddr, err)
	}

	srvOpts := []grpc.ServerOption{}
	unaryServerInterceptors := []grpc.UnaryServerInterceptor{}

	unaryServerInterceptors = append(unaryServerInterceptors, UnaryCancelInterceptor)

	if strings.ToLower(config.Logging.Level) == "debug" {
		unaryServerInterceptors = append(unaryServerInterceptors, UnaryLoggingInterceptor)
	}

	srvOpts = append(srvOpts, grpc.ChainUnaryInterceptor(unaryServerInterceptors...))

	if config.Server.UseTLS {
		creds, err := newServerTLS(&config)
		if err != nil {
			return fmt.Errorf("failed to initiate mtls: %w", err)
		}

		srvOpts = append(srvOpts, grpc.Creds(creds))
	}

	s := grpc.NewServer(srvOpts...)
	reflection.Register(s)
	store.RegisterStoreServer(s, storeServer)

	log.Info().Str("module", "grpc").Str("listen_address", ListenAddr).Msg("gRPC server started")
	return s.Serve(lis)
}

func initLogger(levelStr string) {
	levelStr = strings.ToLower(levelStr)

	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		level = zerolog.InfoLevel
		log.Warn().Str("module", "logging").Msgf("invalid log level '%s' specified, falling back to 'info'", levelStr)
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(level)

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

}
