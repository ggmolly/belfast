package main

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/packets"
	_ "github.com/ggmolly/belfast/internal/protobuf"
)

type packetRecord struct {
	Timestamp string          `json:"ts"`
	Direction string          `json:"dir"`
	StreamID  string          `json:"stream_id"`
	PacketID  int             `json:"packet_id"`
	Length    int             `json:"len"`
	Index     int             `json:"index"`
	JSON      json.RawMessage `json:"json,omitempty"`
	Error     string          `json:"error,omitempty"`
	RawHex    string          `json:"raw_hex,omitempty"`
}

type parseConfig struct {
	ServerPort uint16
	Limit      int
	PacketID   int
	StreamID   string
}

type streamFactory struct {
	cfg          *parseConfig
	registry     map[int]func() proto.Message
	packetCount  int
	outputWriter *bufio.Writer
}

type tcpStream struct {
	net          gopacket.Flow
	transport    gopacket.Flow
	reader       tcpreader.ReaderStream
	cfg          *parseConfig
	registry     map[int]func() proto.Message
	packetCount  *int
	outputWriter *bufio.Writer
}

func (f *streamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	reader := tcpreader.NewReaderStream()
	stream := &tcpStream{
		net:          net,
		transport:    transport,
		reader:       reader,
		cfg:          f.cfg,
		registry:     f.registry,
		packetCount:  &f.packetCount,
		outputWriter: f.outputWriter,
	}
	go stream.run()
	return &stream.reader
}

func (s *tcpStream) run() {
	streamID := fmt.Sprintf("%s:%s", s.net.String(), s.transport.String())
	if s.cfg.StreamID != "" && s.cfg.StreamID != streamID {
		_, _ = io.Copy(io.Discard, &s.reader)
		return
	}
	buffer := make([]byte, 0, 4096)
	tmp := make([]byte, 4096)
	for {
		n, err := s.reader.Read(tmp)
		if n > 0 {
			buffer = append(buffer, tmp[:n]...)
			buffer = s.consumePackets(buffer, streamID)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			logger.WithFields("pcap_decode", logger.FieldValue("stream", streamID)).Error(err.Error())
			return
		}
	}
}

func (s *tcpStream) consumePackets(buffer []byte, streamID string) []byte {
	for {
		if len(buffer) < packets.HEADER_SIZE {
			return buffer
		}
		packetSize := int(binary.BigEndian.Uint16(buffer[0:2]))
		frameSize := packetSize + 2
		if packetSize <= 0 || frameSize > len(buffer) {
			return buffer
		}
		packetID := int(binary.BigEndian.Uint16(buffer[3:5]))
		packetIndex := int(binary.BigEndian.Uint16(buffer[5:7]))
		payload := buffer[packets.HEADER_SIZE:frameSize]
		s.emitPacket(streamID, packetID, packetIndex, payload)
		buffer = buffer[frameSize:]
	}
}

func (s *tcpStream) emitPacket(streamID string, packetID int, packetIndex int, payload []byte) {
	if s.cfg.PacketID != 0 && s.cfg.PacketID != packetID {
		return
	}
	if s.cfg.Limit > 0 && *s.packetCount >= s.cfg.Limit {
		return
	}
	*s.packetCount++

	direction := "CS"
	_, dstPort := streamPorts(s.transport)
	if dstPort != s.cfg.ServerPort {
		direction = "SC"
	}

	record := packetRecord{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Direction: direction,
		StreamID:  streamID,
		PacketID:  packetID,
		Length:    len(payload),
		Index:     packetIndex,
	}
	if constructor, ok := s.registry[packetID]; ok {
		msg := constructor()
		if err := proto.Unmarshal(payload, msg); err != nil {
			record.Error = err.Error()
			record.RawHex = hex.EncodeToString(payload)
		} else {
			marshaled, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(msg)
			if err != nil {
				record.Error = err.Error()
				record.RawHex = hex.EncodeToString(payload)
			} else {
				record.JSON = marshaled
			}
		}
	} else {
		record.Error = "unknown packet id"
		record.RawHex = hex.EncodeToString(payload)
	}
	payloadLine, _ := json.Marshal(record)
	_, _ = s.outputWriter.Write(payloadLine)
	_, _ = s.outputWriter.WriteString("\n")
	_ = s.outputWriter.Flush()
}

func streamPorts(transport gopacket.Flow) (uint16, uint16) {
	src, dst := transport.Endpoints()
	srcPort := layers.TCPPort(binary.BigEndian.Uint16(src.Raw()))
	dstPort := layers.TCPPort(binary.BigEndian.Uint16(dst.Raw()))
	return uint16(srcPort), uint16(dstPort)
}

func buildRegistry() map[int]func() proto.Message {
	registry := map[int]func() proto.Message{}
	protoregistry.GlobalTypes.RangeMessages(func(desc protoreflect.MessageType) bool {
		name := string(desc.Descriptor().Name())
		if !strings.HasPrefix(name, "CS_") && !strings.HasPrefix(name, "SC_") {
			return true
		}
		id, err := strconv.Atoi(name[3:])
		if err != nil {
			return true
		}
		registry[id] = func() proto.Message {
			return desc.New().Interface()
		}
		return true
	})
	return registry
}

func sortedPackets(registry map[int]func() proto.Message) []int {
	keys := make([]int, 0, len(registry))
	for key := range registry {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	return keys
}

func findPacketByName(name string) (int, error) {
	if name == "" {
		return 0, nil
	}
	if len(name) < 3 {
		return 0, fmt.Errorf("invalid packet name %q", name)
	}
	if name[:3] != "CS_" && name[:3] != "SC_" {
		return 0, fmt.Errorf("invalid packet name %q", name)
	}
	var id int
	_, err := fmt.Sscanf(name[3:], "%d", &id)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("invalid packet name %q", name)
	}
	return id, nil
}

func main() {
	pcapPath := flag.String("pcap", "", "path to pcap or pcapng file")
	serverPort := flag.Int("server-port", 0, "server port to infer CS/SC direction")
	limit := flag.Int("limit", 0, "max packets to emit")
	packetFilter := flag.Int("packet", 0, "packet id to filter")
	packetName := flag.String("packet-name", "", "packet name to filter (CS_12002)")
	streamFilter := flag.String("stream", "", "stream id to filter (ip:port-ip:port)")
	flag.Parse()

	if *pcapPath == "" || *serverPort == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "pcap and server-port are required\n")
		os.Exit(2)
	}
	if _, err := os.Stat(*pcapPath); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "pcap not found: %v\n", err)
		os.Exit(2)
	}

	packetID := *packetFilter
	if packetID == 0 && *packetName != "" {
		parsed, err := findPacketByName(*packetName)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(2)
		}
		packetID = parsed
	}

	file, err := os.Open(*pcapPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to open pcap: %v\n", err)
		os.Exit(2)
	}
	defer file.Close()

	var packetReader gopacket.PacketDataSource
	var linkType layers.LinkType
	reader, err := pcapgo.NewReader(file)
	if err != nil {
		if _, errPcapng := file.Seek(0, io.SeekStart); errPcapng == nil {
			if ngReader, errNg := pcapgo.NewNgReader(file, pcapgo.DefaultNgReaderOptions); errNg == nil {
				packetReader = ngReader
				linkType = ngReader.LinkType()
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "failed to read pcap: %v\n", err)
				os.Exit(2)
			}
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "failed to read pcap: %v\n", err)
			os.Exit(2)
		}
	} else {
		packetReader = reader
		linkType = reader.LinkType()
	}

	cfg := &parseConfig{
		ServerPort: uint16(*serverPort),
		Limit:      *limit,
		PacketID:   packetID,
		StreamID:   *streamFilter,
	}

	registry := buildRegistry()
	assembler := tcpassembly.NewAssembler(tcpassembly.NewStreamPool(&streamFactory{
		cfg:          cfg,
		registry:     registry,
		outputWriter: bufio.NewWriter(os.Stdout),
	}))

	source := gopacket.NewPacketSource(packetReader, linkType)
	for packet := range source.Packets() {
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp := tcpLayer.(*layers.TCP)
			netLayer := packet.NetworkLayer()
			if netLayer == nil {
				continue
			}
			assembler.AssembleWithTimestamp(netLayer.NetworkFlow(), tcp, packet.Metadata().Timestamp)
		}
	}
	assembler.FlushAll()

	_, _ = fmt.Fprintf(os.Stderr, "decoded packets (ids=%v) from %s\n", sortedPackets(registry), filepath.Base(*pcapPath))
}

var _ protoreflect.Message
