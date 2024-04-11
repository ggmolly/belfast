package connection

import (
	"bytes"
	"fmt"
	"net"
	"syscall"

	"github.com/bettercallmolly/belfast/logger"
	"github.com/bettercallmolly/belfast/orm"
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
