package node

import (
	"context"
	"io/ioutil"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

// websocketTransport WebSocket传输层结构体
// 继承自loopingTransport，用于处理WebSocket连接的消息循环
type websocketTransport struct {
	*loopingTransport
}

// newWebsocketTransport 创建新的WebSocket传输层实例
// ctx 上下文对象，用于控制连接的生命周期，可通过取消上下文来关闭连接
// addr WebSocket服务器地址
// 返回传输层实例和可能的错误
func newWebsocketTransport(ctx context.Context, addr *url.URL) (transport, error) {
	// 使用默认的WebSocket拨号器建立连接
	wsConn, _, err := websocket.DefaultDialer.DialContext(ctx, addr.String(), nil)
	if err != nil {
		return nil, err
	}

	// 定义读取消息的函数
	readMessage := func() (payload []byte, err error) {
		// 获取下一个消息的读取器
		typ, r, err := wsConn.NextReader()
		if err != nil {
			return nil, errors.Wrap(err, "error reading from backend websocket connection")
		}

		// 只处理文本消息
		if typ != websocket.TextMessage {
			return nil, nil
		}

		// 读取消息内容
		payload, err = ioutil.ReadAll(r)
		if err != nil {
			return nil, errors.Wrap(err, "error reading from backend websocket connection")
		}

		return payload, err
	}

	// 定义写入消息的函数
	writeMessage := func(payload []byte) error {
		err := wsConn.WriteMessage(websocket.TextMessage, payload)
		return err
	}

	// 创建WebSocket传输层实例
	t := websocketTransport{
		loopingTransport: newLoopingTransport(ctx, wsConn, readMessage, writeMessage),
	}

	return &t, nil
}
