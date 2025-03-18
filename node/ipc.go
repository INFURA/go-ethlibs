package node

import (
	"bufio"
	"context"
	"net"
	"net/url"

	"github.com/pkg/errors"
)

// newIPCTransport 创建新的IPC传输层实例
// ctx 上下文对象，用于控制连接的生命周期
// parsedURL 已解析的URL对象，包含Unix域套接字的路径
// 返回IPC传输层实例和可能的错误
func newIPCTransport(ctx context.Context, parsedURL *url.URL) (*ipcTransport, error) {
	// 创建网络拨号器并建立Unix域套接字连接
	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", parsedURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "could not connect over IPC")
	}

	// 创建缓冲扫描器用于读取消息
	scanner := bufio.NewScanner(conn)

	// 定义读取消息的函数
	readMessage := func() (payload []byte, err error) {
		// 扫描下一行数据
		if !scanner.Scan() {
			return nil, ctx.Err()
		}

		// 检查扫描过程中是否有错误
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		// 获取扫描到的文本内容
		payload = []byte(scanner.Text())
		err = nil
		return
	}

	// 定义写入消息的函数
	writeMessage := func(payload []byte) error {
		_, err := conn.Write(payload)
		return err
	}

	// 创建IPC传输层实例
	t := ipcTransport{
		loopingTransport: newLoopingTransport(ctx, conn, readMessage, writeMessage),
	}

	return &t, nil
}

// ipcTransport IPC传输层结构体
// 继承自loopingTransport，用于处理IPC连接的消息循环
type ipcTransport struct {
	*loopingTransport
}
