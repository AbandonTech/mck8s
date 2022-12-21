package tcp_server

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
)

// TcpServer provides a callback handler for on connection
type TcpServer struct {
	Port         int
	OnConnection func(net.Conn)
}

// NewTcpServer creates a TcpServer using the provided port and onConnection method
func NewTcpServer(port int, onConnection func(conn net.Conn)) *TcpServer {
	return &TcpServer{
		port,
		onConnection,
	}
}

// Run the TcpServer in a blocking fashion, serving clients until ctx is closed
func (tcpServer *TcpServer) Run(ctx context.Context) error {
	log.Info().
		Int("Port", tcpServer.Port).
		Msg("Starting TcpServer")

	connectionString := fmt.Sprintf("0.0.0.0:%d", tcpServer.Port)
	listener, err := net.Listen("tcp4", connectionString)
	if err != nil {
		log.Fatal().Err(err)
	}

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Unable to close listener")
		}
	}(listener)

	for {
		log.Debug().
			Msg("Waiting for next connection")
		connection, err := listener.Accept()
		log.Debug().
			Msg("Accepted connection")

		if err != nil {
			log.Error().
				Err(err).
				Msg("Error while accepting connection")
			continue
		}

		go func() {
			defer func(connection net.Conn) {
				err := connection.Close()
				if err != nil {
					log.Error().
						Err(err).
						Msg("Error while closing connection")
				}
			}(connection)

			tcpServer.OnConnection(connection)
		}()
	}
}
