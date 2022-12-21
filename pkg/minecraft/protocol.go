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
	PacketSize       uint64
	PacketId         uint64
	ProtocolVersion  uint64
	ServerAddrLength uint64
	ServerAdd        string
	ServerPort       uint16
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

	packetSize, _ := binary.ReadUvarint(protocolReader.reader)
	packetId, _ := binary.ReadUvarint(protocolReader.reader)

	if packetId != 0 {
		log.Info().Msg("Received non-handshake packet")
		return nil, errors.New("not a handshake packet")
	}

	protocolVer, _ := binary.ReadUvarint(protocolReader.reader)
	serverAddrLength, _ := binary.ReadUvarint(protocolReader.reader)

	var serverAddress = make([]byte, int(serverAddrLength))
	protocolReader.reader.Read(serverAddress)

	var serverPort = make([]byte, 2)
	protocolReader.reader.Read(serverPort)

	return &HandshakePacket{
		packetSize,
		packetId,
		protocolVer,
		serverAddrLength,
		string(serverAddress),
		binary.BigEndian.Uint16(serverPort),
	}, nil
}
