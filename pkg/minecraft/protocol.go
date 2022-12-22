package minecraft

import (
	"bufio"
	"encoding/binary"
	"errors"
	"github.com/rs/zerolog/log"
	"net"
)

// HandshakePacket contains all data in the first handshake packet
type HandshakePacket struct {
	PacketSize       int64
	PacketId         int64
	ProtocolVersion  int64
	ServerAddrLength int64
	ServerAddr       []byte
	ServerPort       []byte
	NextState        uint64
}

// Packet is a sum type of all packets that can be received
type Packet = HandshakePacket

// ProtocolReader wraps a net.Conn to provide an interface to read Packet(s) of data from the client
type ProtocolReader struct {
	reader *bufio.Reader
	con    *net.Conn
}

// NewProtocolReader creates a ProtocolReader from a net.Conn Tcp client
func NewProtocolReader(conn net.Conn) ProtocolReader {
	reader := bufio.NewReader(conn)

	return ProtocolReader{
		reader,
		&conn,
	}
}

// Read a Packet from the net.Conn Tcp stream
func (protocolReader *ProtocolReader) Read() (*Packet, error) {

	packetSize, _ := binary.ReadVarint(protocolReader.reader)
	packetId, _ := binary.ReadVarint(protocolReader.reader)

	if packetId != 0 {
		log.Info().
			Int64("PacketId", packetId).
			Msg("Received non-handshake packet")
		return nil, errors.New("not a handshake packet")
	}

	protocolVer, _ := binary.ReadVarint(protocolReader.reader)
	serverAddrLength, _ := binary.ReadVarint(protocolReader.reader)

	var serverAddress = make([]byte, int(serverAddrLength))
	protocolReader.reader.Read(serverAddress)

	var serverPort = make([]byte, 2)
	protocolReader.reader.Read(serverPort)

	nextState, _ := binary.ReadUvarint(protocolReader.reader)

	return &HandshakePacket{
		packetSize,
		packetId,
		protocolVer,
		serverAddrLength,
		serverAddress,
		serverPort,
		nextState,
	}, nil
}
