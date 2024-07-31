package collector_client

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func AddToGRPCHeader(ctx context.Context, key, val string) context.Context {
	if _, exists := metadata.FromOutgoingContext(ctx); exists {
		return metadata.AppendToOutgoingContext(ctx, key, val)
	}

	return metadata.NewOutgoingContext(ctx, metadata.Pairs(key, val))
}
