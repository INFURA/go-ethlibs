package node

import (
	"context"

	"github.com/ConsenSys/go-ethlibs/jsonrpc"
)

type contextKey string

func ContextWithRequestID(parent context.Context, id jsonrpc.ID) context.Context {
	return context.WithValue(parent, contextKey("id"), &id)
}

func requestIDFromContext(ctx context.Context) *jsonrpc.ID {
	if id, ok := ctx.Value(contextKey("id")).(*jsonrpc.ID); ok {
		return id
	}

	return nil
}
