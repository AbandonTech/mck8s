package tcp

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
)

// TcpServer provides a callback handler for on connection
type TcpServer struct {
	Port         int
	OnConnection func(*TcpServer, net.Conn)
	routes       map[string]string
}

// NewTcpServer creates a TcpServer using the provided port and onConnection method
func NewTcpServer(port int, onConnection func(server *TcpServer, conn net.Conn)) *TcpServer {
	return &TcpServer{
		port,
		onConnection,
		make(map[string]string),
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

			tcpServer.OnConnection(tcpServer, connection)
		}()
	}
}

// AddRoute to the TcpServer
func (tcpServer *TcpServer) AddRoute(hostname string, destination string) {
	if value, ok := tcpServer.routes[hostname]; ok {
		if value != destination {
			log.Warn().
				Str("Hostname", hostname).
				Str("OldRoute", value).
				Str("NewRoute", destination).
				Msg("Overwriting previously created route")
		}
	} else {
		log.Info().
			Str("Hostname", hostname).
			Str("Route", destination).
			Msg("Added new route")
	}

	tcpServer.routes[hostname] = destination
	log.Debug().
		Interface("Routes", tcpServer.routes).
		Msg("Routes updated")
}

// DeleteRoute from the TcpServer
func (tcpServer *TcpServer) DeleteRoute(hostname string) {
	log.Info().
		Str("Hostname", hostname).
		Msg("Removing route")

	delete(tcpServer.routes, hostname)
}

// GetRoute from the TcpServer provides the Service Ip within the Cluster
func (tcpServer *TcpServer) GetRoute(addr string) (string, error) {
	value, ok := tcpServer.routes[addr]
	if !ok {
		return "", errors.New("address is not currently routed")
	}

	return value, nil
}
