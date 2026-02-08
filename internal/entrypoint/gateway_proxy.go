package entrypoint

import (
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/logger"
)

type gatewayProxyUpstreamConfig struct {
	remote                string
	dialTimeout           time.Duration
	requirePrivateClients bool
}

type gatewayProxyRuntime struct {
	bindAddress string
	port        int

	upstream atomic.Value
}

func newGatewayProxyRuntime(cfg config.GatewayConfig) *gatewayProxyRuntime {
	r := &gatewayProxyRuntime{
		bindAddress: cfg.BindAddress,
		port:        cfg.Port,
	}
	r.Update(cfg)
	return r
}

func (r *gatewayProxyRuntime) Update(cfg config.GatewayConfig) {
	requirePrivate := true
	if cfg.RequirePrivateClients != nil {
		requirePrivate = *cfg.RequirePrivateClients
	}
	r.upstream.Store(gatewayProxyUpstreamConfig{
		remote:                cfg.ProxyRemote,
		dialTimeout:           time.Duration(cfg.ProxyDialTimeoutMS) * time.Millisecond,
		requirePrivateClients: requirePrivate,
	})
}

func (r *gatewayProxyRuntime) Run() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", r.bindAddress, r.port))
	if err != nil {
		logger.LogEvent("Gateway", "Proxy", fmt.Sprintf("error listening: %v", err), logger.LOG_LEVEL_ERROR)
		return err
	}
	defer listener.Close()
	logger.LogEvent("Gateway", "Proxy", fmt.Sprintf("proxy listening on %s:%d", r.bindAddress, r.port), logger.LOG_LEVEL_INFO)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.LogEvent("Gateway", "Proxy", fmt.Sprintf("error accepting: %v", err), logger.LOG_LEVEL_ERROR)
			continue
		}
		go r.handle(conn)
	}
}

func (r *gatewayProxyRuntime) handle(client net.Conn) {
	upstreamCfgAny := r.upstream.Load()
	upstreamCfg, ok := upstreamCfgAny.(gatewayProxyUpstreamConfig)
	if !ok {
		_ = client.Close()
		return
	}

	if upstreamCfg.requirePrivateClients {
		if tcpAddr, ok := client.RemoteAddr().(*net.TCPAddr); ok {
			if tcpAddr.IP != nil && !tcpAddr.IP.IsPrivate() {
				logger.LogEvent("Gateway", "Proxy", fmt.Sprintf("rejecting non-private client %s", client.RemoteAddr().String()), logger.LOG_LEVEL_INFO)
				_ = client.Close()
				return
			}
		}
	}

	if upstreamCfg.remote == "" {
		logger.LogEvent("Gateway", "Proxy", "proxy_remote is empty", logger.LOG_LEVEL_ERROR)
		_ = client.Close()
		return
	}

	dialer := net.Dialer{
		Timeout:   upstreamCfg.dialTimeout,
		KeepAlive: 30 * time.Second,
	}
	upstream, err := dialer.Dial("tcp", upstreamCfg.remote)
	if err != nil {
		logger.LogEvent("Gateway", "Proxy", fmt.Sprintf("dial %s failed: %v", upstreamCfg.remote, err), logger.LOG_LEVEL_ERROR)
		_ = client.Close()
		return
	}

	logger.WithFields(
		"Gateway",
		logger.FieldValue("mode", "proxy"),
		logger.FieldValue("client", client.RemoteAddr().String()),
		logger.FieldValue("upstream", upstreamCfg.remote),
	).Info("proxy connection established")

	clientToUpstream, upstreamToClient := proxyBidirectional(client, upstream)
	logger.WithFields(
		"Gateway",
		logger.FieldValue("mode", "proxy"),
		logger.FieldValue("client", client.RemoteAddr().String()),
		logger.FieldValue("upstream", upstreamCfg.remote),
		logger.FieldValue("bytes_client_to_upstream", clientToUpstream),
		logger.FieldValue("bytes_upstream_to_client", upstreamToClient),
	).Info("proxy connection closed")
}

func proxyBidirectional(client net.Conn, upstream net.Conn) (int64, int64) {
	var closeOnce sync.Once
	closeBoth := func() {
		closeOnce.Do(func() {
			_ = client.Close()
			_ = upstream.Close()
		})
	}

	var wg sync.WaitGroup
	wg.Add(2)

	var clientToUpstream int64
	var upstreamToClient int64

	go func() {
		defer wg.Done()
		buf := make([]byte, 32<<10)
		n, _ := io.CopyBuffer(upstream, client, buf)
		clientToUpstream = n
		closeBoth()
	}()

	go func() {
		defer wg.Done()
		buf := make([]byte, 32<<10)
		n, _ := io.CopyBuffer(client, upstream, buf)
		upstreamToClient = n
		closeBoth()
	}()

	wg.Wait()
	return clientToUpstream, upstreamToClient
}
