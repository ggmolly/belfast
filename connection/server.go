package connection

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"sync"

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
	rooms       map[uint32][]*Client
	Region      string
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
	return &client, err
}

func (server *Server) AddClient(client *Client) {
	logger.LogEvent("Server", "hewwo", fmt.Sprintf("new connection from %s:%d", client.IP, client.Port), logger.LOG_LEVEL_DEBUG)
}

func (server *Server) RemoveClient(client *Client) {
	logger.LogEvent("Server", "cya", fmt.Sprintf("%s:%d", client.IP, client.Port), logger.LOG_LEVEL_DEBUG)
	(*client.Connection).Close()
}

func handleConnection(conn net.Conn, wg *sync.WaitGroup, server *Server) {
	defer wg.Done()
	defer conn.Close()

	// Add the client to the list
	client, err := server.GetClient(&conn)

	if err != nil {
		logger.LogEvent("Server", "Handler", fmt.Sprintf("client %s -- error: %v", conn.RemoteAddr(), err), logger.LOG_LEVEL_ERROR)
		conn.Close()
		server.RemoveClient(client)
		return
	}

	if !client.IP.IsPrivate() {
		logger.LogEvent("Server", "Handler", fmt.Sprintf("client %s -- not in a private range", conn.RemoteAddr()), logger.LOG_LEVEL_ERROR)
		conn.Close()
		server.RemoveClient(client)
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
			// We have a full message, slice it and send it to the dispatcher
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

func (server *Server) Run() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.BindAddress, server.Port))
	if err != nil {
		logger.LogEvent("Server", "Run", fmt.Sprintf("error listening: %v", err), logger.LOG_LEVEL_ERROR)
		return
	}
	defer listener.Close()
	logger.LogEvent("Server", "Run", fmt.Sprintf("listening on %s:%d", server.BindAddress, server.Port), logger.LOG_LEVEL_INFO)

	var wg sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.LogEvent("Server", "Run", fmt.Sprintf("error accepting: %v", err), logger.LOG_LEVEL_ERROR)
			continue
		}
		wg.Add(1)
		go handleConnection(conn, &wg, server)
	}
	wg.Wait()
}

func NewServer(bindAddress string, port int, dispatcher ServerDispatcher) *Server {
	return &Server{
		BindAddress: bindAddress,
		Port:        port,
		Dispatcher:  dispatcher,
		Region:      os.Getenv("AL_REGION"),
		rooms:       make(map[uint32][]*Client),
	}
}

// Chat room management
func (server *Server) JoinRoom(roomID uint32, client *Client) {
	server.rooms[roomID] = append(server.rooms[roomID], client)
}

func (server *Server) LeaveRoom(roomID uint32, client *Client) {
	for i, c := range server.rooms[roomID] {
		if c == client {
			server.rooms[roomID] = append(server.rooms[roomID][:i], server.rooms[roomID][i+1:]...)
			break
		}
	}
}

func (server *Server) ChangeRoom(oldRoomID uint32, newRoomID uint32, client *Client) {
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
