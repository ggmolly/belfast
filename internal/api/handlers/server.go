package handlers

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/region"
)

type ServerHandler struct {
	Config *config.Config
}

func RegisterServerRoutes(party iris.Party, handler *ServerHandler) {
	party.Get("/status", handler.Status)
	party.Post("/start", handler.Start)
	party.Post("/stop", handler.Stop)
	party.Post("/restart", handler.Restart)
	party.Get("/config", handler.GetConfig)
	party.Put("/config", handler.UpdateConfig)
	party.Get("/stats", handler.Stats)
	party.Get("/metrics", handler.Metrics)
	party.Get("/connections", handler.ListConnections)
	party.Get("/connections/{id}", handler.ConnectionDetail)
	party.Delete("/connections/{id}", handler.DisconnectConnection)
	party.Get("/uptime", handler.Uptime)
}

func (handler *ServerHandler) Status(ctx iris.Context) {
	server := connection.BelfastInstance
	uptime := time.Since(server.StartTime)
	payload := types.ServerStatusResponse{
		Running:     true,
		Accepting:   server.IsAcceptingConnections(),
		UptimeSec:   int64(uptime.Seconds()),
		UptimeHuman: uptime.String(),
		ClientCount: server.ClientCount(),
	}
	_ = ctx.JSON(response.Success(payload))
}

func (handler *ServerHandler) Start(ctx iris.Context) {
	connection.BelfastInstance.SetAcceptingConnections(true)
	_ = ctx.JSON(response.Success(nil))
}

func (handler *ServerHandler) Stop(ctx iris.Context) {
	connection.BelfastInstance.SetAcceptingConnections(false)
	_ = ctx.JSON(response.Success(nil))
}

func (handler *ServerHandler) Restart(ctx iris.Context) {
	server := connection.BelfastInstance
	server.SetAcceptingConnections(false)
	server.SetAcceptingConnections(true)
	_ = ctx.JSON(response.Success(nil))
}

func (handler *ServerHandler) GetConfig(ctx iris.Context) {
	cfg := handler.Config
	payload := types.ServerConfigResponse{
		BindAddress: cfg.Belfast.BindAddress,
		Port:        cfg.Belfast.Port,
		Region:      cfg.Region.Default,
	}
	_ = ctx.JSON(response.Success(payload))
}

func (handler *ServerHandler) UpdateConfig(ctx iris.Context) {
	var req types.ServerConfigUpdate
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if err := validateConfigUpdate(req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	cfg := handler.Config
	cfg.Belfast.BindAddress = req.BindAddress
	cfg.Belfast.Port = req.Port
	cfg.Region.Default = req.Region
	if err := region.SetCurrent(req.Region); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

func (handler *ServerHandler) Stats(ctx iris.Context) {
	payload := types.ServerStatsResponse{
		ClientCount: connection.BelfastInstance.ClientCount(),
	}
	_ = ctx.JSON(response.Success(payload))
}

func (handler *ServerHandler) Metrics(ctx iris.Context) {
	server := connection.BelfastInstance
	clients := server.ListClients()
	metrics := aggregateMetrics(clients)
	payload := types.ServerMetricsResponse{
		ClientCount:   len(clients),
		QueueMax:      metrics.QueueMax,
		QueueBlocks:   metrics.QueueBlocks,
		HandlerErrors: metrics.HandlerErrors,
		WriteErrors:   metrics.WriteErrors,
		PacketsPerSec: metrics.PacketsPerSec,
	}
	_ = ctx.JSON(response.Success(payload))
}

func (handler *ServerHandler) ListConnections(ctx iris.Context) {
	clients := connection.BelfastInstance.ListClients()
	payload := make([]types.ConnectionSummary, 0, len(clients))
	for _, client := range clients {
		payload = append(payload, types.ConnectionSummary{
			Hash:        client.Hash,
			RemoteAddr:  net.JoinHostPort(client.IP.String(), strconv.Itoa(client.Port)),
			ConnectedAt: client.ConnectedAt.Format(time.RFC3339),
			CommanderID: commanderID(client),
		})
	}
	_ = ctx.JSON(response.Success(payload))
}

func (handler *ServerHandler) ConnectionDetail(ctx iris.Context) {
	hash, err := parseHash(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	client, ok := connection.BelfastInstance.FindClient(hash)
	if !ok {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "connection not found", nil))
		return
	}
	stats := client.MetricsSnapshot()
	payload := types.ConnectionDetail{
		Hash:          client.Hash,
		RemoteAddr:    net.JoinHostPort(client.IP.String(), strconv.Itoa(client.Port)),
		ConnectedAt:   client.ConnectedAt.Format(time.RFC3339),
		CommanderID:   commanderID(client),
		QueueMax:      stats.QueueMax,
		QueueBlocks:   stats.QueueBlocks,
		HandlerErrors: stats.HandlerErrors,
		WriteErrors:   stats.WriteErrors,
	}
	_ = ctx.JSON(response.Success(payload))
}

func (handler *ServerHandler) DisconnectConnection(ctx iris.Context) {
	hash, err := parseHash(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	client, ok := connection.BelfastInstance.FindClient(hash)
	if !ok {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "connection not found", nil))
		return
	}
	connection.BelfastInstance.RemoveClient(client)
	_ = ctx.JSON(response.Success(nil))
}

func (handler *ServerHandler) Uptime(ctx iris.Context) {
	uptime := time.Since(connection.BelfastInstance.StartTime)
	payload := types.ServerUptimeResponse{
		UptimeSec:   int64(uptime.Seconds()),
		UptimeHuman: uptime.String(),
	}
	_ = ctx.JSON(response.Success(payload))
}

func validateConfigUpdate(req types.ServerConfigUpdate) error {
	if strings.TrimSpace(req.BindAddress) == "" {
		return fmt.Errorf("bind_address is required")
	}
	if req.Port < 1 || req.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if err := region.Validate(req.Region); err != nil {
		return err
	}
	return nil
}

func parseHash(input string) (uint32, error) {
	parsed, err := strconv.ParseUint(input, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid connection id")
	}
	return uint32(parsed), nil
}

func commanderID(client *connection.Client) uint32 {
	if client.Commander == nil {
		return 0
	}
	return client.Commander.CommanderID
}

type metricsTotals struct {
	QueueMax      int
	QueueBlocks   uint64
	HandlerErrors uint64
	WriteErrors   uint64
	PacketsPerSec float64
}

func aggregateMetrics(clients []*connection.Client) metricsTotals {
	var totals metricsTotals
	var packetCount uint64
	for _, client := range clients {
		stats := client.MetricsSnapshot()
		if stats.QueueMax > totals.QueueMax {
			totals.QueueMax = stats.QueueMax
		}
		totals.QueueBlocks += stats.QueueBlocks
		totals.HandlerErrors += stats.HandlerErrors
		totals.WriteErrors += stats.WriteErrors
		packetCount += stats.Packets
	}
	if len(clients) > 0 {
		totals.PacketsPerSec = float64(packetCount)
	}
	return totals
}
