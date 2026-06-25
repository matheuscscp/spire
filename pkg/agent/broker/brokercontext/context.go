package brokercontext

import (
	"context"
	"errors"
	"fmt"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"google.golang.org/grpc/metadata"
)

const (
	callerIDMetadataKey       = "spire-agent-broker-caller-id"
	callerNodeNameMetadataKey = "spire-agent-broker-caller-node-name"
)

type callerIDKey struct{}
type callerNodeNameKey struct{}

func WithCallerID(ctx context.Context, id spiffeid.ID) context.Context {
	return context.WithValue(ctx, callerIDKey{}, id)
}

func WithCallerNodeName(ctx context.Context, nodeName string) context.Context {
	if nodeName == "" {
		return ctx
	}
	return context.WithValue(ctx, callerNodeNameKey{}, nodeName)
}

func CallerIDFromContext(ctx context.Context) (spiffeid.ID, bool, error) {
	if id, ok := ctx.Value(callerIDKey{}).(spiffeid.ID); ok {
		return id, true, nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return spiffeid.ID{}, false, nil
	}
	values := md.Get(callerIDMetadataKey)
	switch len(values) {
	case 0:
		return spiffeid.ID{}, false, nil
	case 1:
		id, err := spiffeid.FromString(values[0])
		if err != nil {
			return spiffeid.ID{}, false, fmt.Errorf("invalid broker caller SPIFFE ID: %w", err)
		}
		return id, true, nil
	default:
		return spiffeid.ID{}, false, errors.New("multiple broker caller SPIFFE IDs provided")
	}
}

func CallerNodeNameFromContext(ctx context.Context) (string, bool, error) {
	if nodeName, ok := ctx.Value(callerNodeNameKey{}).(string); ok {
		return nodeName, nodeName != "", nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false, nil
	}
	values := md.Get(callerNodeNameMetadataKey)
	switch len(values) {
	case 0:
		return "", false, nil
	case 1:
		if values[0] == "" {
			return "", false, nil
		}
		return values[0], true, nil
	default:
		return "", false, errors.New("multiple broker caller node names provided")
	}
}

func AppendCallerIDToOutgoingContext(ctx context.Context) context.Context {
	if id, ok := ctx.Value(callerIDKey{}).(spiffeid.ID); ok {
		ctx = metadata.AppendToOutgoingContext(ctx, callerIDMetadataKey, id.String())
	}
	if nodeName, ok := ctx.Value(callerNodeNameKey{}).(string); ok && nodeName != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, callerNodeNameMetadataKey, nodeName)
	}
	return ctx
}
