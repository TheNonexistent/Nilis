package main

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func UnaryCancelInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	select {
	case <-ctx.Done():
		err := ctx.Err()

		p, ok := peer.FromContext(ctx)
		if ok {
			log.Warn().
				Str("module", "grpc").
				Str("method", info.FullMethod).
				Str("peer", p.Addr.String()).
				Err(err).
				Msg("context cancelled before handler")
		} else {
			log.Warn().
				Str("module", "grpc").
				Str("method", info.FullMethod).
				Err(err).
				Msg("context cancelled before handler")
		}

		return nil, status.Errorf(codes.Canceled, "request cancelled: %v", err)
	default:
	}

	return handler(ctx, req)
}

func UnaryLoggingInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	if p, ok := peer.FromContext(ctx); ok {
		log.Debug().
			Str("module", "grpc").
			Str("method", info.FullMethod).
			Str("peer", p.Addr.String()).
			Interface("request", req).
			Msg("handling request")
	} else {
		log.Debug().
			Str("module", "grpc").
			Str("method", info.FullMethod).
			Interface("request", req).
			Msg("handling request")
	}

	return handler(ctx, req)
}
