package connection

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"
	queue "gopkg.in/eapache/queue.v1"

	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	rngutil "github.com/ggmolly/belfast/internal/rng"
)

const (
	packetQueueSize = 512
	packetPoolSize  = 128
)

var (
	ErrClientClosed = errors.New("client is closed")

	accountIdRandom = rngutil.NewLockedRand()
)

type ClientMetrics struct {
	queueMax      int
	queueBlocks   uint64
	handlerErrors uint64
	writeErrors   uint64
	packets       uint64
}

type Client struct {
	IP              net.IP
	Port            int
	State           int
	PacketIndex     int
	Hash            uint32
	Connection      *net.Conn
	Commander       *orm.Commander
	AuthArg2        uint32
	Buffer          bytes.Buffer
	Server          *Server
	ConnectedAt     time.Time
	PreviousLoginAt time.Time

	packetQueue  *queue.Queue
	queueMu      sync.Mutex
	queueCond    *sync.Cond
	queueLimit   int
	packetPool   chan []byte
	closed       bool
	closeOnce    sync.Once
	dispatchOnce sync.Once
	metrics      ClientMetrics
}

func (client *Client) initQueues() {
	client.queueLimit = packetQueueSize
	client.packetQueue = queue.New()
	client.queueCond = sync.NewCond(&client.queueMu)
	client.packetPool = make(chan []byte, packetPoolSize)
}

func (client *Client) StartDispatcher() {
	client.dispatchOnce.Do(func() {
		go client.dispatchLoop()
	})
}

func (client *Client) EnqueuePacket(packet []byte) error {
	client.queueMu.Lock()
	defer client.queueMu.Unlock()
	for client.packetQueue.Length() >= client.queueLimit && !client.closed {
		client.metrics.queueBlocks++
		logger.LogEvent("Metrics", "QueueBlock", fmt.Sprintf("%s:%d depth=%d", client.IP, client.Port, client.packetQueue.Length()), logger.LOG_LEVEL_DEBUG)
		client.queueCond.Wait()
	}
	if client.closed {
		return ErrClientClosed
	}
	client.packetQueue.Add(packet)
	depth := client.packetQueue.Length()
	if depth > client.metrics.queueMax {
		client.metrics.queueMax = depth
		logger.LogEvent("Metrics", "QueueDepth", fmt.Sprintf("%s:%d depth=%d", client.IP, client.Port, depth), logger.LOG_LEVEL_DEBUG)
	}
	client.queueCond.Signal()
	return nil
}

func (client *Client) Close() {
	client.CloseWithError(nil)
}

func (client *Client) CloseWithError(err error) {
	client.closeOnce.Do(func() {
		client.queueMu.Lock()
		client.closed = true
		if client.queueCond != nil {
			client.queueCond.Broadcast()
		}
		client.queueMu.Unlock()
		if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, ErrClientClosed) {
			logger.LogEvent("Client", "Close", fmt.Sprintf("%s:%d -> %v", client.IP, client.Port, err), logger.LOG_LEVEL_ERROR)
		}
		if client.Connection != nil {
			_ = (*client.Connection).Close()
		}
		client.logMetrics()
	})
}

func (client *Client) IsClosed() bool {
	client.queueMu.Lock()
	closed := client.closed
	client.queueMu.Unlock()
	return closed
}

func (client *Client) RecordHandlerError() {
	atomic.AddUint64(&client.metrics.handlerErrors, 1)
}

func (client *Client) recordWriteError() {
	atomic.AddUint64(&client.metrics.writeErrors, 1)
}

func (client *Client) logMetrics() {
	client.queueMu.Lock()
	queueMax := client.metrics.queueMax
	queueBlocks := client.metrics.queueBlocks
	client.queueMu.Unlock()
	handlerErrors := atomic.LoadUint64(&client.metrics.handlerErrors)
	writeErrors := atomic.LoadUint64(&client.metrics.writeErrors)
	packets := atomic.LoadUint64(&client.metrics.packets)
	logger.LogEvent("Metrics", "ClientStats", fmt.Sprintf("%s:%d queueMax=%d queueBlocks=%d handlerErrors=%d writeErrors=%d packets=%d", client.IP, client.Port, queueMax, queueBlocks, handlerErrors, writeErrors, packets), logger.LOG_LEVEL_INFO)
}

func (client *Client) acquirePacketBuffer(size int) []byte {
	select {
	case buf := <-client.packetPool:
		if cap(buf) >= size {
			return buf[:size]
		}
	default:
	}
	return make([]byte, size)
}

func (client *Client) releasePacketBuffer(buffer []byte) {
	if buffer == nil {
		return
	}
	select {
	case client.packetPool <- buffer[:0]:
	default:
	}
}

func (client *Client) drainQueueLocked() {
	for client.packetQueue.Length() > 0 {
		packet := client.packetQueue.Remove().([]byte)
		client.releasePacketBuffer(packet)
	}
}

func (client *Client) dispatchLoop() {
	for {
		client.queueMu.Lock()
		for client.packetQueue.Length() == 0 && !client.closed {
			client.queueCond.Wait()
		}
		if client.closed {
			client.drainQueueLocked()
			client.queueMu.Unlock()
			return
		}
		packet := client.packetQueue.Remove().([]byte)
		client.queueCond.Signal()
		client.queueMu.Unlock()

		if client.IsClosed() {
			client.releasePacketBuffer(packet)
			return
		}
		atomic.AddUint64(&client.metrics.packets, 1)
		client.Server.Dispatcher(&packet, client, len(packet))
		client.releasePacketBuffer(packet)
	}
}

func (client *Client) CreateCommander(arg2 uint32) (uint32, error) {
	accountId := accountIdRandom.Uint32()
	if accountId == 0 {
		accountId = 1
	}
	// Tie an account to passed arg2 (which is some sort of account identifier)
	if err := orm.GormDB.Create(&orm.YostarusMap{
		Arg2:      arg2,
		AccountID: accountId,
	}).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to create account for arg2 %d: %v", arg2, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}
	// Create a new commander for the account
	if err := orm.GormDB.Create(&orm.Commander{
		AccountID:   accountId,
		CommanderID: accountId,
		Name:        fmt.Sprintf("Unnamed commander #%d", accountId),
		GuideIndex:  1,
		// TODO: Confirm initial new guide index once guide versioning is finalized.
		NewGuideIndex: 1,
	}).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to create commander for account %d: %v", accountId, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}

	// Since we have no tutorial / first login, we'll also give a secretary to the new commander
	belfast := orm.OwnedShip{
		OwnerID:           accountId,
		ShipID:            202124, // Belfast (6 stars)
		IsSecretary:       true,
		SecretaryPosition: proto.Uint32(0),
	}
	if err := orm.GormDB.Create(&belfast).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to give Belfast to account %d: %v", accountId, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}

	longIsland := orm.OwnedShip{
		OwnerID: accountId,
		ShipID:  106011, // Long Island
	}
	if err := orm.GormDB.Create(&longIsland).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to give Long Island to account %d: %v", accountId, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}

	// Give default items to commander
	if err := orm.GormDB.Create(&([]orm.CommanderItem{{
		// Wisdom Cube
		CommanderID: accountId,
		ItemID:      20001,
		Count:       1,
	}, {
		// Quick Finisher
		CommanderID: accountId,
		ItemID:      15003,
		Count:       10,
	}})).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to give default items to account %d: %v", accountId, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}
	// Give default resources to commander
	if err := orm.GormDB.Create(&([]orm.OwnedResource{{
		// Gold
		CommanderID: accountId,
		ResourceID:  1,
		Amount:      3000,
	}, {
		// Oil
		CommanderID: accountId,
		ResourceID:  2,
		Amount:      500,
	}, {
		// Gem
		CommanderID: accountId,
		ResourceID:  4,
		Amount:      0,
	}})).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to give default resources to account %d: %v", accountId, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}

	if err := orm.GormDB.Create(&orm.Fleet{
		CommanderID:    accountId,
		GameID:         1,
		Name:           "",
		ShipList:       orm.Int64List{int64(belfast.ID), int64(longIsland.ID)},
		MeowfficerList: orm.Int64List{},
	}).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to create default fleet for account %d: %v", accountId, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}

	logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("created new commander for account %d", accountId), logger.LOG_LEVEL_INFO)
	return accountId, nil
}

func (client *Client) CreateCommanderWithStarter(arg2 uint32, nickname string, shipID uint32) (uint32, error) {
	accountId := accountIdRandom.Uint32()
	if accountId == 0 {
		accountId = 1
	}
	if err := orm.GormDB.Create(&orm.YostarusMap{
		Arg2:      arg2,
		AccountID: accountId,
	}).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to create account for arg2 %d: %v", arg2, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}
	if err := orm.GormDB.Create(&orm.Commander{
		AccountID:   accountId,
		CommanderID: accountId,
		Name:        nickname,
		GuideIndex:  1,
		// TODO: Confirm initial new guide index once guide versioning is finalized.
		NewGuideIndex: 1,
	}).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to create commander for account %d: %v", accountId, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}
	starterShip := orm.OwnedShip{
		OwnerID: accountId,
		ShipID:  shipID,
	}
	if err := orm.GormDB.Create(&starterShip).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to give starter ship to account %d: %v", accountId, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}
	belfast := orm.OwnedShip{
		OwnerID:           accountId,
		ShipID:            202124, // Belfast (6 stars)
		IsSecretary:       true,
		SecretaryPosition: proto.Uint32(0),
	}
	if err := orm.GormDB.Create(&belfast).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to give Belfast to account %d: %v", accountId, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}
	longIsland := orm.OwnedShip{
		OwnerID: accountId,
		ShipID:  106011, // Long Island
	}
	if err := orm.GormDB.Create(&longIsland).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to give Long Island to account %d: %v", accountId, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}
	if err := orm.GormDB.Create(&([]orm.CommanderItem{{
		CommanderID: accountId,
		ItemID:      20001,
		Count:       1,
	}, {
		CommanderID: accountId,
		ItemID:      15003,
		Count:       10,
	}})).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to give default items to account %d: %v", accountId, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}
	if err := orm.GormDB.Create(&([]orm.OwnedResource{{
		CommanderID: accountId,
		ResourceID:  1,
		Amount:      3000,
	}, {
		CommanderID: accountId,
		ResourceID:  2,
		Amount:      500,
	}, {
		CommanderID: accountId,
		ResourceID:  4,
		Amount:      0,
	}})).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to give default resources to account %d: %v", accountId, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}
	if err := orm.GormDB.Create(&orm.Fleet{
		CommanderID:    accountId,
		GameID:         1,
		Name:           "",
		ShipList:       orm.Int64List{int64(starterShip.ID), int64(belfast.ID), int64(longIsland.ID)},
		MeowfficerList: orm.Int64List{},
	}).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to create default fleet for account %d: %v", accountId, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}
	logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("created new commander for account %d", accountId), logger.LOG_LEVEL_INFO)
	return accountId, nil
}

func (client *Client) GetCommander(accountId uint32) error {
	err := orm.GormDB.Where("account_id = ?", accountId).First(&client.Commander).Error
	return err
}

// Sends SC_10999 (disconnected from server) message to the Client, reasons are defined in consts/disconnect_reasons.go
func (client *Client) Disconnect(reason uint8) error {
	_, _, err := SendProtoMessage(10999, client, &protobuf.SC_10999{
		Reason: proto.Uint32(uint32(reason)),
	})
	return err
}

// Sends the content of the buffer to the client via TCP
func (client *Client) Flush() error {
	_, err := (*client.Connection).Write(client.Buffer.Bytes())
	if err != nil {
		client.recordWriteError()
		logger.LogEvent("Client", "Flush", fmt.Sprintf("%s:%d -> %v", client.IP, client.Port, err), logger.LOG_LEVEL_ERROR)
		client.Buffer.Reset()
		client.CloseWithError(err)
		return err
	}
	client.Buffer.Reset()
	return nil
}

func (client *Client) SendMessage(packetId int, message any) (int, int, error) {
	return SendProtoMessage(packetId, client, message)
}

type MetricsSnapshot struct {
	QueueMax      int
	QueueBlocks   uint64
	HandlerErrors uint64
	WriteErrors   uint64
	Packets       uint64
}

func (client *Client) MetricsSnapshot() MetricsSnapshot {
	client.queueMu.Lock()
	queueMax := client.metrics.queueMax
	queueBlocks := client.metrics.queueBlocks
	client.queueMu.Unlock()
	return MetricsSnapshot{
		QueueMax:      queueMax,
		QueueBlocks:   queueBlocks,
		HandlerErrors: atomic.LoadUint64(&client.metrics.handlerErrors),
		WriteErrors:   atomic.LoadUint64(&client.metrics.writeErrors),
		Packets:       atomic.LoadUint64(&client.metrics.packets),
	}
}
