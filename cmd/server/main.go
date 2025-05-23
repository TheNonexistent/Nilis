package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/thenonexistent/nilis/pkg/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

var config Config

func main() {

	if err := loadConfig(&config); err != nil {
		log.Fatal().Str("module", "main").Err(err).Msg("error in initial configuration")
	}

	if err := validateConfig(&config); err != nil {
		log.Fatal().Str("module", "main").Err(err).Msg("invalid configuration")
	}

	initLogger(config.Logging.Level)

	storeServer, cancelFunc, err := NewServer(&config)
	if err != nil {
		log.Fatal().Str("module", "main").Err(err).Msg("failed to create store server")
	}
	defer cancelFunc()

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
		creds, err := credentials.NewServerTLSFromFile(config.Server.TLSCert, config.Server.TLSKey)
		if err != nil {
			return fmt.Errorf("failed to load certificates: %w", err)
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
