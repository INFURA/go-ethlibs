package node

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/INFURA/go-ethlibs/jsonrpc"
	"github.com/pkg/errors"
)

// newHTTPTransport 创建新的HTTP传输层实例
// ctx 上下文对象，用于控制请求的生命周期
// parsedURL 已解析的URL对象，包含目标节点的地址信息
// 返回传输层实例和可能的错误
func newHTTPTransport(ctx context.Context, parsedURL *url.URL) (transport, error) {
	return &httpTransport{
		rawURL: parsedURL.String(),
	}, nil
}

// httpTransport HTTP传输层结构体
type httpTransport struct {
	rawURL string       // 目标节点的原始URL字符串
	client *http.Client // HTTP客户端实例
	once   sync.Once    // 确保客户端只初始化一次的同步控制
}

// Request 处理JSON-RPC请求
// ctx 上下文对象，用于控制请求的生命周期
// r 待处理的JSON-RPC请求
// 返回JSON-RPC响应和可能的错误
func (t *httpTransport) Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error) {
	// 将请求对象序列化为JSON字节数组
	b, err := json.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode request json")
	}

	// 发送请求并获取响应数据
	body, err := t.dispatchBytes(ctx, b)
	if err != nil {
		return nil, errors.Wrap(err, "could not dispatch request")
	}

	// 将响应数据解析为JSON-RPC响应对象
	jr := jsonrpc.RawResponse{}
	err = json.Unmarshal(body, &jr)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode response json")
	}

	return &jr, nil
}

// Subscribe 处理订阅请求（HTTP不支持订阅）
// ctx 上下文对象
// r JSON-RPC请求
// 始终返回错误，因为HTTP协议不支持订阅功能
func (t *httpTransport) Subscribe(ctx context.Context, r *jsonrpc.Request) (Subscription, error) {
	return nil, errors.New("subscriptions not supported over HTTP")
}

// IsBidirectional 检查传输层是否支持双向通信
// 返回false，因为HTTP协议是单向的，不支持服务器推送
func (t *httpTransport) IsBidirectional() bool {
	return false
}

// dispatchBytes 发送HTTP请求并获取响应数据
// ctx 上下文对象，用于控制请求的生命周期
// input 请求体数据
// 返回响应体数据和可能的错误
func (t *httpTransport) dispatchBytes(ctx context.Context, input []byte) ([]byte, error) {
	// 使用sync.Once确保HTTP客户端只初始化一次
	t.once.Do(func() {
		// 由于此客户端只用于访问单个端点，
		// 我们允许所有空闲连接指向该主机
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.MaxIdleConnsPerHost = tr.MaxIdleConns
		t.client = &http.Client{
			Timeout:   120 * time.Second, // 设置120秒超时
			Transport: tr,
		}
	})

	// 创建新的HTTP请求
	r, err := http.NewRequest(http.MethodPost, t.rawURL, bytes.NewReader(input))
	if err != nil {
		return nil, errors.Wrap(err, "could not create http.Request")
	}

	// 设置请求上下文和内容类型
	r = r.WithContext(ctx)
	r.Header.Add("Content-Type", "application/json")

	// 发送请求并获取响应
	resp, err := t.client.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "error in client.Do")
	}

	defer resp.Body.Close()

	// 读取响应体数据
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading body")
	}

	return body, nil
}
