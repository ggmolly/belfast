package handlers

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/authz"
	"github.com/ggmolly/belfast/internal/buildinfo"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/region"
)

type ServerHandler struct {
	Config *config.Config
}

func RegisterServerRoutes(party iris.Party, handler *ServerHandler) {
	party.Get("/status", handler.Status)
	party.Post("/start", middleware.RequirePermissionAny(authz.PermServer), handler.Start)
	party.Post("/stop", middleware.RequirePermissionAny(authz.PermServer), handler.Stop)
	party.Post("/restart", middleware.RequirePermissionAny(authz.PermServer), handler.Restart)
	party.Post("/maintenance", middleware.RequirePermissionAny(authz.PermServer), handler.Maintenance)
	party.Get("/maintenance", middleware.RequirePermissionAny(authz.PermServer), handler.MaintenanceStatus)
	party.Get("/config", middleware.RequirePermissionAny(authz.PermServer), handler.GetConfig)
	party.Put("/config", middleware.RequirePermissionAny(authz.PermServer), handler.UpdateConfig)
	party.Get("/stats", middleware.RequirePermissionAny(authz.PermServer), handler.Stats)
	party.Get("/metrics", middleware.RequirePermissionAny(authz.PermServer), handler.Metrics)
	party.Get("/connections", middleware.RequirePermissionAny(authz.PermServer), handler.ListConnections)
	party.Get("/connections/{id}", middleware.RequirePermissionAny(authz.PermServer), handler.ConnectionDetail)
	party.Delete("/connections/{id}", middleware.RequirePermissionAny(authz.PermServer), handler.DisconnectConnection)
	party.Get("/uptime", middleware.RequirePermissionAny(authz.PermServer), handler.Uptime)
}

// Status godoc
// @Summary     Get server status
// @Tags        Server
// @Produce     json
// @Success     200  {object}  ServerStatusResponseDoc
// @Router      /api/v1/server/status [get]
func (handler *ServerHandler) Status(ctx iris.Context) {
	server := connection.BelfastInstance
	uptime := time.Since(server.StartTime)
	payload := types.ServerStatusResponse{
		Name:        handler.Config.Belfast.Name,
		Commit:      buildinfo.ShortCommit(),
		Running:     true,
		Accepting:   server.IsAcceptingConnections(),
		Maintenance: server.MaintenanceEnabled(),
		UptimeSec:   int64(uptime.Seconds()),
		UptimeHuman: uptime.String(),
		ClientCount: server.ClientCount(),
	}
	_ = ctx.JSON(response.Success(payload))
}

// Start godoc
// @Summary     Start accepting connections
// @Tags        Server
// @Produce     json
// @Success     200  {object}  OKResponseDoc
// @Router      /api/v1/server/start [post]
func (handler *ServerHandler) Start(ctx iris.Context) {
	connection.BelfastInstance.SetAcceptingConnections(true)
	_ = ctx.JSON(response.Success(nil))
}

// Stop godoc
// @Summary     Stop accepting connections
// @Tags        Server
// @Produce     json
// @Success     200  {object}  OKResponseDoc
// @Router      /api/v1/server/stop [post]
func (handler *ServerHandler) Stop(ctx iris.Context) {
	connection.BelfastInstance.SetAcceptingConnections(false)
	_ = ctx.JSON(response.Success(nil))
}

// Restart godoc
// @Summary     Restart accepting connections
// @Tags        Server
// @Produce     json
// @Success     200  {object}  OKResponseDoc
// @Router      /api/v1/server/restart [post]
func (handler *ServerHandler) Restart(ctx iris.Context) {
	server := connection.BelfastInstance
	server.SetAcceptingConnections(false)
	server.SetAcceptingConnections(true)
	_ = ctx.JSON(response.Success(nil))
}

// Maintenance godoc
// @Summary     Toggle maintenance mode
// @Tags        Server
// @Accept      json
// @Produce     json
// @Param       body  body  types.ServerMaintenanceUpdate  true  "Maintenance toggle"
// @Success     200  {object}  ServerMaintenanceResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/server/maintenance [post]
func (handler *ServerHandler) Maintenance(ctx iris.Context) {
	var req types.ServerMaintenanceUpdate
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if err := handler.Config.PersistMaintenance(req.Enabled); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to persist maintenance", nil))
		return
	}
	connection.BelfastInstance.SetMaintenance(req.Enabled)
	payload := types.ServerMaintenanceResponse{Enabled: connection.BelfastInstance.MaintenanceEnabled()}
	_ = ctx.JSON(response.Success(payload))
}

// MaintenanceStatus godoc
// @Summary     Get maintenance status
// @Tags        Server
// @Produce     json
// @Success     200  {object}  ServerMaintenanceResponseDoc
// @Router      /api/v1/server/maintenance [get]
func (handler *ServerHandler) MaintenanceStatus(ctx iris.Context) {
	payload := types.ServerMaintenanceResponse{Enabled: connection.BelfastInstance.MaintenanceEnabled()}
	_ = ctx.JSON(response.Success(payload))
}

// GetConfig godoc
// @Summary     Get server config
// @Tags        Server
// @Produce     json
// @Success     200  {object}  ServerConfigResponseDoc
// @Router      /api/v1/server/config [get]
func (handler *ServerHandler) GetConfig(ctx iris.Context) {
	cfg := handler.Config
	payload := types.ServerConfigResponse{
		BindAddress: cfg.Belfast.BindAddress,
		Port:        cfg.Belfast.Port,
		Region:      cfg.Region.Default,
	}
	_ = ctx.JSON(response.Success(payload))
}

// UpdateConfig godoc
// @Summary     Update server config
// @Tags        Server
// @Accept      json
// @Produce     json
// @Param       payload  body  types.ServerConfigUpdate  true  "Server config"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Router      /api/v1/server/config [put]
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

// Stats godoc
// @Summary     Get server stats
// @Tags        Server
// @Produce     json
// @Success     200  {object}  ServerStatsResponseDoc
// @Router      /api/v1/server/stats [get]
func (handler *ServerHandler) Stats(ctx iris.Context) {
	payload := types.ServerStatsResponse{
		ClientCount: connection.BelfastInstance.ClientCount(),
	}
	_ = ctx.JSON(response.Success(payload))
}

// Metrics godoc
// @Summary     Get server metrics
// @Tags        Server
// @Produce     json
// @Success     200  {object}  ServerMetricsResponseDoc
// @Router      /api/v1/server/metrics [get]
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

// ListConnections godoc
// @Summary     List active connections
// @Tags        Server
// @Produce     json
// @Success     200  {object}  ConnectionListResponseDoc
// @Router      /api/v1/server/connections [get]
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

// ConnectionDetail godoc
// @Summary     Get connection details
// @Tags        Server
// @Produce     json
// @Param       id   path  int  true  "Connection ID"
// @Success     200  {object}  ConnectionDetailResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Router      /api/v1/server/connections/{id} [get]
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

// DisconnectConnection godoc
// @Summary     Disconnect connection
// @Tags        Server
// @Produce     json
// @Param       id   path  int  true  "Connection ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Router      /api/v1/server/connections/{id} [delete]
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

// Uptime godoc
// @Summary     Get server uptime
// @Tags        Server
// @Produce     json
// @Success     200  {object}  ServerUptimeResponseDoc
// @Router      /api/v1/server/uptime [get]
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
