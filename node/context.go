// Package node 提供了以太坊节点交互相关的功能
package node

import (
	"context"

	"github.com/INFURA/go-ethlibs/jsonrpc"
)

// contextKey 是用于在context中存储值的自定义键类型
type contextKey string

// ContextWithRequestID 创建一个新的上下文，其中包含指定的请求ID
// parent: 父上下文
// id: 要存储的JSON-RPC请求ID
// 返回: 包含请求ID的新上下文
func ContextWithRequestID(parent context.Context, id jsonrpc.ID) context.Context {
	return context.WithValue(parent, contextKey("id"), &id)
}

// requestIDFromContext 从上下文中获取请求ID
// ctx: 包含请求ID的上下文
// 返回: 如果上下文中存在请求ID则返回该ID的指针，否则返回nil
func requestIDFromContext(ctx context.Context) *jsonrpc.ID {
	if id, ok := ctx.Value(contextKey("id")).(*jsonrpc.ID); ok {
		return id
	}

	return nil
}
