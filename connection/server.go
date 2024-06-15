package connection

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"sync"

	"github.com/ggmolly/belfast/consts"
	"github.com/ggmolly/belfast/debug"
	"github.com/ggmolly/belfast/logger"
	"github.com/ggmolly/belfast/orm"
	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

type ServerDispatcher func(*[]byte, *Client, int)

type Server struct {
	BindAddress string
	Port        int
	SocketFD    int
	EpollFD     int
	Dispatcher  ServerDispatcher
	Region      string

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
	for _, c := range fmt.Sprintf("%s:%d", client.IP, client.Port) {
		client.Hash += uint32(c)
	}
	return &client, err
}

func (server *Server) AddClient(client *Client) {
	logger.LogEvent("Server", "Hello", fmt.Sprintf("new connection from %s:%d", client.IP, client.Port), logger.LOG_LEVEL_DEBUG)
	client.Server.clientsMutex.Lock()
	defer client.Server.clientsMutex.Unlock()
	server.clients[client.Hash] = client
}

func (server *Server) RemoveClient(client *Client) {
	client.Server.clientsMutex.Lock()
	defer client.Server.clientsMutex.Unlock()
	logger.LogEvent("Server", "Goodbye", fmt.Sprintf("%s:%d", client.IP, client.Port), logger.LOG_LEVEL_DEBUG)
	(*client.Connection).Close()
	delete(server.clients, client.Hash)
}

func handleConnection(conn net.Conn, server *Server) {
	logger.LogEvent("Server", "TEST", "Goroutine started", logger.LOG_LEVEL_WARN)
	defer conn.Close()
	defer logger.LogEvent("Server", "TEST", "Goroutine ended", logger.LOG_LEVEL_WARN)
	// Add the client to the list
	client, err := server.GetClient(&conn)

	if err != nil {
		logger.LogEvent("Server", "Handler", fmt.Sprintf("client %s -- error: %v", conn.RemoteAddr(), err), logger.LOG_LEVEL_ERROR)
		conn.Close()
		return
	}

	if !client.IP.IsPrivate() {
		logger.LogEvent("Server", "Handler", fmt.Sprintf("client %s -- not in a private range", conn.RemoteAddr()), logger.LOG_LEVEL_ERROR)
		conn.Close()
		return
	}

	server.AddClient(client)

	// Buffer for unpacking received data
	totalBytes := 0
	packerBuffer := make([]byte, 16384)

	// Temporary buffer for reading
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err == io.EOF || err != nil {
			conn.Close()
			server.RemoveClient(client)
			break
		}
		// copy the buffer to the packerBuffer
		copy(packerBuffer[totalBytes:], buffer[:n])
		totalBytes += n

		// To know if we have atleast a full message, check first 2 bytes of the packerBuffer
		// these two bytes are the length of a message
		size := int(packerBuffer[0])<<8 | int(packerBuffer[1]) + 2 // take into account the 2 bytes for the size
		if totalBytes >= size {
			// Slice the packerBuffer to get the message and send it to the dispatcher
			message := packerBuffer[:size]
			server.Dispatcher(&message, client, size)
			// Remove the message from the packerBuffer and shift the rest of the buffer
			packerBuffer = packerBuffer[size:]
			totalBytes -= size
		} else {
			// Otherwise, wait for more data
			continue
		}
	}
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
		go handleConnection(conn, server)
	}
}

func NewServer(bindAddress string, port int, dispatcher ServerDispatcher) *Server {
	return &Server{
		BindAddress: bindAddress,
		Port:        port,
		Dispatcher:  dispatcher,
		Region:      os.Getenv("AL_REGION"),
		clients:     make(map[uint32]*Client),
		rooms:       make(map[uint32][]*Client),
	}
}

// Sends SC_10999 (disconnected from server) message to every connected clients, reasons are defined in consts/disconnect_reasons.go
func (server *Server) DisconnectAll(reason uint8) {
	server.clientsMutex.Lock()
	defer server.clientsMutex.Unlock()
	for _, client := range server.clients {
		logger.LogEvent("Server", "Disconnect", fmt.Sprintf("disconnecting %s:%d -> %s", client.IP, client.Port, consts.ResolveReason(reason)), logger.LOG_LEVEL_DEBUG)
		client.Disconnect(reason)
		client.Flush()
		(*client.Connection).Close()
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
		Player: &protobuf.PLAYER_INFO{
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
		return 0, packetId, err
	}
	debug.InsertPacket(packetId, &data)
	InjectPacketHeader(packetId, &data, client.PacketIndex)
	n, err := client.Buffer.Write(data)
	logger.LogEvent("Connection", "SendMessage", fmt.Sprintf("SC_%d - %d bytes buffered", packetId, n), logger.LOG_LEVEL_DEBUG)
	return n, packetId, err
}
