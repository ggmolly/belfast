package connection

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/smallnest/ringbuffer"
	"google.golang.org/protobuf/proto"

	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/debug"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/ggmolly/belfast/internal/region"
)

const (
	readBufferSize = 32 << 10
)

type ServerDispatcher func(*[]byte, *Client, int)

type Server struct {
	BindAddress string
	Port        int
	SocketFD    int
	EpollFD     int
	Dispatcher  ServerDispatcher
	Region      string
	StartTime   time.Time

	acceptingConnections atomic.Bool

	maintenanceEnabled uint32

	// Maps & mutexes
	roomsMutex   sync.RWMutex
	rooms        map[uint32][]*Client // Game chat rooms
	clientsMutex sync.RWMutex
	clients      map[uint32]*Client // Socket hash -> Client
}

var (
	BelfastInstance *Server
)

func (server *Server) GetClient(conn *net.Conn) (*Client, error) {
	var client Client
	var err error
	client.IP = (*conn).RemoteAddr().(*net.TCPAddr).IP
	client.Port = (*conn).RemoteAddr().(*net.TCPAddr).Port
	client.Connection = conn
	client.Server = server
	client.ConnectedAt = time.Now().UTC()
	client.initQueues()
	for _, c := range fmt.Sprintf("%s:%d", client.IP, client.Port) {
		client.Hash += uint32(c)
	}
	return &client, err
}

func (server *Server) AddClient(client *Client) {
	logger.LogEvent("Server", "Hello", fmt.Sprintf("new connection from %s:%d", client.IP, client.Port), logger.LOG_LEVEL_DEBUG)
	server.clientsMutex.Lock()
	defer server.clientsMutex.Unlock()
	client.Server = server
	server.clients[client.Hash] = client
}

func (server *Server) RemoveClient(client *Client) {
	server.clientsMutex.Lock()
	defer server.clientsMutex.Unlock()
	logger.LogEvent("Server", "Goodbye", fmt.Sprintf("%s:%d", client.IP, client.Port), logger.LOG_LEVEL_DEBUG)
	client.Close()
	delete(server.clients, client.Hash)
}

func (server *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()
	// Add the client to the list
	client, err := server.GetClient(&conn)
	if err == nil {
		logger.WithFields("Server", logger.FieldValue("remote", conn.RemoteAddr().String()), logger.FieldValue("local", conn.LocalAddr().String()), logger.FieldValue("private", client.IP.IsPrivate())).Info("connection accepted")
	}

	if err != nil {
		logger.LogEvent("Server", "Handler", fmt.Sprintf("client %s -- error: %v", conn.RemoteAddr(), err), logger.LOG_LEVEL_ERROR)
		conn.Close()
		return
	}

	if server.MaintenanceEnabled() {
		logger.LogEvent("Server", "Run", fmt.Sprintf("maintenance enabled, rejecting %s", conn.RemoteAddr().String()), logger.LOG_LEVEL_INFO)
		conn.Close()
		return
	}

	if !client.IP.IsPrivate() {
		logger.WithFields("Server", logger.FieldValue("remote", conn.RemoteAddr().String()), logger.FieldValue("local", conn.LocalAddr().String())).Error("client not in private range")
		conn.Close()
		return
	}
	if !server.IsAcceptingConnections() {
		logger.LogEvent("Server", "Reject", fmt.Sprintf("rejecting %s:%d (stopped)", client.IP, client.Port), logger.LOG_LEVEL_INFO)
		conn.Close()
		return
	}

	server.AddClient(client)
	client.StartDispatcher()

	ring := ringbuffer.New(readBufferSize).SetBlocking(true)
	go func() {
		_, err := ring.ReadFrom(conn)
		if err != nil && !errors.Is(err, io.EOF) {
			logger.LogEvent("Server", "Read", fmt.Sprintf("%s:%d -> %v", client.IP, client.Port, err), logger.LOG_LEVEL_ERROR)
		}
		ring.CloseWriter()
	}()

	for {
		if client.IsClosed() {
			server.RemoveClient(client)
			return
		}
		sizeHeader, size, err := readPacketSize(ring)
		if err == nil {
			logger.WithFields("Server", logger.FieldValue("remote", conn.RemoteAddr().String()), logger.FieldValue("size", size)).Debug("read packet size")
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				logger.LogEvent("Server", "Read", fmt.Sprintf("%s:%d -> %v", client.IP, client.Port, err), logger.LOG_LEVEL_ERROR)
			}
			server.RemoveClient(client)
			return
		}
		if size == 0 {
			continue
		}
		packet := client.acquirePacketBuffer(size)
		copy(packet[:2], sizeHeader[:])
		logger.WithFields("Server", logger.FieldValue("remote", conn.RemoteAddr().String()), logger.FieldValue("size", size)).Debug("reading packet body")
		if err := readPacketBody(ring, packet[2:]); err != nil {
			client.releasePacketBuffer(packet)
			if !errors.Is(err, io.EOF) {
				logger.LogEvent("Server", "Read", fmt.Sprintf("%s:%d -> %v", client.IP, client.Port, err), logger.LOG_LEVEL_ERROR)
			}
			server.RemoveClient(client)
			return
		}
		if err := client.EnqueuePacket(packet); err != nil {
			client.releasePacketBuffer(packet)
			server.RemoveClient(client)
			return
		}
	}
}

func readPacketSize(ring *ringbuffer.RingBuffer) ([2]byte, int, error) {
	var header [2]byte
	if _, err := io.ReadFull(ring, header[:]); err != nil {
		return header, 0, err
	}
	size := int(header[0])<<8 | int(header[1])
	if size < 5 {
		return header, 0, fmt.Errorf("invalid packet size %d", size)
	}
	return header, size + 2, nil
}

func readPacketBody(ring *ringbuffer.RingBuffer, packet []byte) error {
	if len(packet) == 0 {
		return nil
	}
	_, err := io.ReadFull(ring, packet)
	return err
}

func (server *Server) Run() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.BindAddress, server.Port))
	if err != nil {
		logger.LogEvent("Server", "Run", fmt.Sprintf("error listening: %v", err), logger.LOG_LEVEL_ERROR)
		return err
	}
	defer listener.Close()
	logger.LogEvent("Server", "Run", fmt.Sprintf("listening on %s:%d", server.BindAddress, server.Port), logger.LOG_LEVEL_INFO)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.LogEvent("Server", "Run", fmt.Sprintf("error accepting: %v", err), logger.LOG_LEVEL_ERROR)
			continue
		}
		go server.HandleConnection(conn)
	}
}

func (server *Server) SetAcceptingConnections(enabled bool) {
	server.acceptingConnections.Store(enabled)
	if !enabled {
		server.DisconnectAll(consts.DR_CONNECTION_TO_SERVER_LOST)
	}
}

func (server *Server) IsAcceptingConnections() bool {
	return server.acceptingConnections.Load()
}

func (server *Server) ClientCount() int {
	server.clientsMutex.RLock()
	defer server.clientsMutex.RUnlock()
	return len(server.clients)
}

func (server *Server) ListClients() []*Client {
	server.clientsMutex.RLock()
	defer server.clientsMutex.RUnlock()
	clients := make([]*Client, 0, len(server.clients))
	for _, client := range server.clients {
		clients = append(clients, client)
	}
	return clients
}

func (server *Server) FindClient(hash uint32) (*Client, bool) {
	server.clientsMutex.RLock()
	defer server.clientsMutex.RUnlock()
	client, ok := server.clients[hash]
	return client, ok
}

func (server *Server) FindClientByCommander(commanderID uint32) (*Client, bool) {
	server.clientsMutex.RLock()
	defer server.clientsMutex.RUnlock()
	for _, client := range server.clients {
		if client.Commander != nil && client.Commander.CommanderID == commanderID {
			return client, true
		}
	}
	return nil, false
}

func (server *Server) DisconnectCommander(commanderID uint32, reason uint8, excludeClient *Client) bool {
	server.clientsMutex.Lock()
	defer server.clientsMutex.Unlock()

	var existingClient *Client
	for _, client := range server.clients {
		if client.Commander != nil && client.Commander.CommanderID == commanderID {
			existingClient = client
			break
		}
	}
	if existingClient == nil {
		return false
	}
	if excludeClient != nil && existingClient == excludeClient {
		return false
	}

	logger.LogEvent("Server", "LoginKick",
		fmt.Sprintf("kicking commander %d from %s:%d",
			commanderID, existingClient.IP, existingClient.Port),
		logger.LOG_LEVEL_INFO)

	existingClient.Disconnect(reason)
	if existingClient.Connection != nil {
		if err := existingClient.Flush(); err != nil {
			logger.LogEvent("Server", "LoginKick",
				fmt.Sprintf("failed to flush %s:%d -> %v", existingClient.IP, existingClient.Port, err),
				logger.LOG_LEVEL_ERROR)
		}
	}
	existingClient.Close()
	delete(server.clients, existingClient.Hash)
	return true
}

func NewServer(bindAddress string, port int, dispatcher ServerDispatcher) *Server {
	server := &Server{
		BindAddress:        bindAddress,
		Port:               port,
		Dispatcher:         dispatcher,
		Region:             region.Current(),
		StartTime:          time.Now(),
		maintenanceEnabled: 0,
		clients:            make(map[uint32]*Client),
		rooms:              make(map[uint32][]*Client),
	}
	server.acceptingConnections.Store(true)
	BelfastInstance = server
	return server
}

func (server *Server) SetMaintenance(enabled bool) {
	value := uint32(0)
	if enabled {
		value = 1
	}
	atomic.StoreUint32(&server.maintenanceEnabled, value)
	if enabled {
		server.DisconnectAll(consts.DR_SERVER_MAINTENANCE)
	}
}

func (server *Server) MaintenanceEnabled() bool {
	return atomic.LoadUint32(&server.maintenanceEnabled) == 1
}

// Sends SC_10999 (disconnected from server) message to every connected clients, reasons are defined in consts/disconnect_reasons.go
func (server *Server) DisconnectAll(reason uint8) {
	server.clientsMutex.Lock()
	defer server.clientsMutex.Unlock()
	for _, client := range server.clients {
		logger.LogEvent("Server", "Disconnect", fmt.Sprintf("disconnecting %s:%d -> %s", client.IP, client.Port, consts.ResolveReason(reason)), logger.LOG_LEVEL_DEBUG)
		client.Disconnect(reason)
		if err := client.Flush(); err != nil {
			logger.LogEvent("Server", "Disconnect", fmt.Sprintf("failed to flush %s:%d -> %v", client.IP, client.Port, err), logger.LOG_LEVEL_ERROR)
		}
		client.Close()
		delete(server.clients, client.Hash)
	}
}

// Chat room management
func (server *Server) JoinRoom(roomID uint32, client *Client) {
	server.roomsMutex.Lock()
	defer server.roomsMutex.Unlock()
	server.rooms[roomID] = append(server.rooms[roomID], client)
}

func (server *Server) LeaveRoom(roomID uint32, client *Client) {
	server.roomsMutex.Lock()
	defer server.roomsMutex.Unlock()
	for i, c := range server.rooms[roomID] {
		if c == client {
			server.rooms[roomID] = append(server.rooms[roomID][:i], server.rooms[roomID][i+1:]...)
			break
		}
	}
}

func (server *Server) ChangeRoom(oldRoomID uint32, newRoomID uint32, client *Client) {
	server.roomsMutex.Lock()
	defer server.roomsMutex.Unlock()
	for i, c := range server.rooms[oldRoomID] {
		if c == client {
			server.rooms[oldRoomID] = append(server.rooms[oldRoomID][:i], server.rooms[oldRoomID][i+1:]...)
			break
		}
	}
	server.rooms[newRoomID] = append(server.rooms[newRoomID], client)
}

func (server *Server) SendMessage(sender *Client, message orm.Message) {
	msgPacket := protobuf.SC_50101{
		Player: &protobuf.PLAYER_INFO_P50{
			Id:   proto.Uint32(sender.Commander.CommanderID),
			Name: proto.String(sender.Commander.Name),
			Lv:   proto.Uint32(uint32(sender.Commander.Level)),
		},
		Type:    proto.Uint32(orm.MSG_TYPE_NORMAL),
		Content: proto.String(message.Content),
	}
	server.roomsMutex.RLock()
	defer server.roomsMutex.RUnlock()
	for _, client := range server.rooms[message.RoomID] {
		client.SendMessage(50101, &msgPacket)
	}
}

func (server *Server) BroadcastGuildChat(message *protobuf.SC_60008) {
	server.clientsMutex.RLock()
	defer server.clientsMutex.RUnlock()
	for _, client := range server.clients {
		client.SendMessage(60008, message)
	}
}

func GeneratePacketHeader(packetId int, payload *[]byte, packetIndex int) []byte {
	var buffer bytes.Buffer

	payloadSize := len(*payload) + 5
	buffer.Write([]byte{byte(payloadSize >> 8), byte(payloadSize)})
	buffer.Write([]byte{0x00})
	buffer.Write([]byte{byte(packetId >> 8), byte(packetId)})
	buffer.Write([]byte{byte(packetIndex >> 8), byte(packetIndex)})

	return buffer.Bytes()
}

func InjectPacketHeader(packetId int, payload *[]byte, packetIndex int) {
	// prepend the header
	header := GeneratePacketHeader(packetId, payload, packetIndex)
	*payload = append(header, *payload...)
}

func SendProtoMessage(packetId int, client *Client, message any) (int, int, error) {
	if reflect.TypeOf(message).Kind() != reflect.Ptr {
		return 0, packetId, fmt.Errorf("message must be a pointer")
	}
	if !reflect.TypeOf(message).Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) {
		return 0, packetId, fmt.Errorf("message must be a proto.Message")
	}
	data, err := proto.Marshal(message.(proto.Message))
	if err != nil {
		logger.LogEvent("Connection", "Marshal", fmt.Sprintf("SC_%d -> %v", packetId, err), logger.LOG_LEVEL_ERROR)
		client.RecordHandlerError()
		client.CloseWithError(err)
		return 0, packetId, err
	}
	debug.InsertPacket(packetId, &data)
	InjectPacketHeader(packetId, &data, client.PacketIndex)
	n, err := client.Buffer.Write(data)
	if err != nil {
		logger.LogEvent("Connection", "Buffer", fmt.Sprintf("SC_%d -> %v", packetId, err), logger.LOG_LEVEL_ERROR)
		client.RecordHandlerError()
		client.CloseWithError(err)
		return n, packetId, err
	}
	logger.LogEvent("Connection", "SendMessage", fmt.Sprintf("SC_%d - %d bytes buffered", packetId, n), logger.LOG_LEVEL_DEBUG)
	return n, packetId, nil
}
