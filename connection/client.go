package connection

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"syscall"

	"github.com/ggmolly/belfast/logger"
	"github.com/ggmolly/belfast/orm"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	SockAddr    syscall.Sockaddr
	ProxyFD     int // only used in proxy strategy, contains the fd of the proxy client
	IP          net.IP
	Port        int
	FD          int
	State       int
	PacketIndex int
	Commander   *orm.Commander
	Buffer      bytes.Buffer
	Server      *Server
}

func (client *Client) CreateCommander(arg2 uint32) (uint32, error) {
	accountId := uint32(rand.Uint32())
	if accountId == 0 {
		accountId = 1
	}
	// Tie an account to passed arg2 (which is some sort of account identifier)
	if err := orm.GormDB.Create(&orm.YostarusMap{Arg2: arg2, AccountID: accountId}).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to create account for arg2 %d: %v", arg2, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}
	// Create a new commander for the account
	if err := orm.GormDB.Create(&orm.Commander{
		AccountID:   accountId,
		CommanderID: accountId,
		Name:        fmt.Sprintf("Unnamed commander #%d", accountId),
	}).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to create commander for account %d: %v", accountId, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}

	// Since we have no tutorial / first login, we'll also give a secretary to the new commander
	if err := orm.GormDB.Create(&orm.OwnedShip{
		OwnerID:           accountId,
		ShipID:            202124, // Belfast (6 stars)
		IsSecretary:       true,
		SecretaryPosition: proto.Uint32(0),
	}).Error; err != nil {
		logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("failed to give Belfast to account %d: %v", accountId, err), logger.LOG_LEVEL_ERROR)
		return 0, err
	}
	logger.LogEvent("Client", "CreateCommander", fmt.Sprintf("created new commander for account %d", accountId), logger.LOG_LEVEL_INFO)
	return accountId, nil
}

func (client *Client) GetCommander(accountId uint32) error {
	err := orm.GormDB.Where("account_id = ?", accountId).First(&client.Commander).Error
	return err
}

func (client *Client) Kill() {
	if err := syscall.EpollCtl(client.Server.EpollFD, syscall.EPOLL_CTL_DEL, client.FD, nil); err != nil {
		logger.LogEvent("Client", "Kill()", fmt.Sprintf("%s:%d -> %v", client.IP, client.Port, err), logger.LOG_LEVEL_ERROR)
	}
	if err := syscall.Close(client.FD); err != nil {
		logger.LogEvent("Client", "Kill()", fmt.Sprintf("%s:%d -> %v", client.IP, client.Port, err), logger.LOG_LEVEL_ERROR)
		return
	}
}

// Sends the content of the buffer to the client via TCP
func (client *Client) Flush() {
	_, err := syscall.Write(client.FD, client.Buffer.Bytes())
	if err != nil {
		logger.LogEvent("Client", "Flush()", fmt.Sprintf("%s:%d -> %v", client.IP, client.Port, err), logger.LOG_LEVEL_ERROR)
	}
	// logger.LogEvent("Client", "Flush()", fmt.Sprintf("%s:%d -> %d bytes flushed", client.IP, client.Port, n), logger.LOG_LEVEL_INFO)
	client.Buffer.Reset()
}

func (client *Client) SendMessage(packetId int, message any) (int, int, error) {
	return SendProtoMessage(packetId, client, message)
}
