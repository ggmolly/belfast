package connection

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"reflect"
	"syscall"

	"github.com/ggmolly/belfast/debug"
	"github.com/ggmolly/belfast/logger"
	"github.com/ggmolly/belfast/orm"
	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

type ServerDispatcher func(*[]byte, *Client)

type Server struct {
	BindAddress string
	Port        int
	SocketFD    int
	EpollFD     int
	Clients     map[int]*Client
	Dispatcher  ServerDispatcher
	rooms       map[uint32][]*Client
	Region      string
}

var (
	BelfastInstance *Server
)

func (server *Server) GetClient(fd int) (*Client, error) {
	var client Client
	var err error
	client.SockAddr, err = syscall.Getpeername(fd)
	if err != nil {
		return &client, err
	}
	client.IP = client.SockAddr.(*syscall.SockaddrInet4).Addr[:]
	client.Port = client.SockAddr.(*syscall.SockaddrInet4).Port
	client.FD = fd
	client.Server = server
	return &client, nil
}

func (server *Server) GetConnectedClient(fd int) (*Client, error) {
	if client, ok := server.Clients[fd]; ok {
		return client, nil
	}
	return nil, fmt.Errorf("client not found")
}

func (server *Server) AddClient(client *Client) {
	logger.LogEvent("Server", "hewwo", fmt.Sprintf("new connection from %s:%d (fd=%d)", client.IP, client.Port, client.FD), logger.LOG_LEVEL_DEBUG)
	server.Clients[client.FD] = client
}

func (server *Server) RemoveClient(client *Client) {
	logger.LogEvent("Server", "cya", fmt.Sprintf("%s:%d (fd=%d)", client.IP, client.Port, client.FD), logger.LOG_LEVEL_DEBUG)
	client.Kill()
	delete(server.Clients, client.FD)
}

func (server *Server) Run() error {
	var err error
	BelfastInstance = server
	if server.SocketFD, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM|syscall.O_NONBLOCK, 0); err != nil {
		return fmt.Errorf("failed to create socket : %v", err)
	}
	defer syscall.Close(server.SocketFD)
	logger.LogEvent("Server", "Listen", fmt.Sprintf("Listening on %s:%d", server.BindAddress, server.Port), logger.LOG_LEVEL_AUTO)

	if err = syscall.SetsockoptInt(server.SocketFD, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return fmt.Errorf("setsockopt error: %v", err)
	}

	if err = syscall.SetNonblock(server.SocketFD, true); err != nil {
		return fmt.Errorf("setnonblock error: %v", err)
	}

	var ip [4]byte
	copy(ip[:], net.ParseIP(server.BindAddress).To4())
	addr := syscall.SockaddrInet4{
		Port: server.Port,
		Addr: ip,
	}

	if err = syscall.Bind(server.SocketFD, &addr); err != nil {
		return fmt.Errorf("bind error: %v", err)
	}

	if err = syscall.Listen(server.SocketFD, syscall.SOMAXCONN); err != nil {
		return fmt.Errorf("listen error: %v", err)
	}

	if server.EpollFD, err = syscall.EpollCreate1(0); err != nil {
		panic(err)
	}

	// Prepare epoll (I/O multiplexing)
	var event syscall.EpollEvent
	event.Events = syscall.EPOLLIN
	event.Fd = int32(server.SocketFD)
	if err = syscall.EpollCtl(server.EpollFD, syscall.EPOLL_CTL_ADD, server.SocketFD, &event); err != nil {
		panic(err)
	}

	// Create epoll event buffer
	var events [128]syscall.EpollEvent
	for {
		// Check for events
		var nevents int
		if nevents, err = syscall.EpollWait(server.EpollFD, events[:], -1); err != nil {
			if err == syscall.EINTR {
				continue
			}
			panic(err)
		}
		var treatedEvents int
		for ev := 0; ev < nevents; ev++ {
			treatedEvents++
			if int(events[ev].Fd) == server.SocketFD {
				// Accept new connections
				connFd, _, err := syscall.Accept(server.SocketFD)
				if err != nil {
					logger.LogEvent("Server", "Accept", fmt.Sprintf("accept error: %v", err), logger.LOG_LEVEL_ERROR)
					continue
				}

				// Make the connection non-blocking
				if err = syscall.SetNonblock(connFd, true); err != nil {
					logger.LogEvent("Server", "SetNonblock", fmt.Sprintf("setnonblock error: %v", err), logger.LOG_LEVEL_ERROR)
					syscall.Close(connFd)
					continue
				}
				event.Events = syscall.EPOLLIN
				event.Fd = int32(connFd)
				if err := syscall.EpollCtl(server.EpollFD, syscall.EPOLL_CTL_ADD, connFd, &event); err != nil {
					logger.LogEvent("Server", "EpollCtl", fmt.Sprintf("epoll_ctl error: %v", err), logger.LOG_LEVEL_ERROR)
					syscall.Close(connFd)
					continue
				}
				// Add the client to the list
				client, err := server.GetClient(connFd)
				if err != nil {
					logger.LogEvent("Server", "GetClient", fmt.Sprintf("getclient error: %v", err), logger.LOG_LEVEL_ERROR)
					continue
				}
				if !client.IP.IsPrivate() {
					logger.LogEvent("Server", "GetClient", fmt.Sprintf("client %s:%d is not in a private range", client.IP, client.Port), logger.LOG_LEVEL_ERROR)
					syscall.EpollCtl(server.EpollFD, syscall.EPOLL_CTL_DEL, connFd, &event)
					syscall.Close(connFd)
					continue
				}
				server.AddClient(client)
			} else {
				// Handle data
				var buffer = make([]byte, 8192)
				clientFd := int(events[ev].Fd)
				client, err := server.GetConnectedClient(clientFd)
				if err != nil {
					logger.LogEvent("Server", "GetConnectedClient", fmt.Sprintf("%v", err), logger.LOG_LEVEL_ERROR)
					server.RemoveClient(client)
					continue
				}
				n, err := syscall.Read(clientFd, buffer)
				if err != nil { // the client probably closed the connection
					logger.LogEvent("Server", "Read", fmt.Sprintf("%v", err), logger.LOG_LEVEL_ERROR)
				} else if n > 0 {
					buffer = buffer[:n]
					if len(buffer) >= 7 {
						server.Dispatcher(&buffer, client)
					}
				} else {
					// EOF, delete from epoll
					server.RemoveClient(client)
				}
			}
		}
		if treatedEvents != nevents {
			panic(fmt.Errorf("treated %d events out of %d", treatedEvents, nevents))
		}
	}
}

func (server *Server) Kill() {
	logger.LogEvent("Server", "Kill()", "Closing server", logger.LOG_LEVEL_INFO)
	if err := syscall.Close(server.SocketFD); err != nil {
		logger.LogEvent("Server", "Kill()", fmt.Sprintf("error closing socket: %v", err), logger.LOG_LEVEL_ERROR)
	}
	if err := syscall.Close(server.EpollFD); err != nil {
		logger.LogEvent("Server", "Kill()", fmt.Sprintf("error closing epoll: %v", err), logger.LOG_LEVEL_ERROR)
	}
	// Close all clients
	for _, client := range server.Clients {
		client.Kill()
	}
}

func NewServer(bindAddress string, port int, dispatcher ServerDispatcher) *Server {
	return &Server{
		BindAddress: bindAddress,
		Port:        port,
		Dispatcher:  dispatcher,
		Clients:     make(map[int]*Client),
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

// TODO: Expose publicly these functions, and delete the package `packets`
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
