package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"minecraftk8s/internal"
	"minecraftk8s/pkg/k8s"
	"minecraftk8s/pkg/tcp"
	"net"
)

func main() {
	internal.ConfigureLogging()
	log.Info().Msg("Starting mck8s-ingress-controller")

	tcpServer := tcp.NewTcpServer(25565, func(server *tcp.TcpServer, clientConn net.Conn) {
		remote := clientConn.RemoteAddr()
		log.Info().
			Str("Client", remote.String()).
			Msg("Connection received")

		packet := make([]byte, 100)
		clientConn.Read(packet)

		packetReader := bufio.NewReader(bytes.NewReader(packet))
		_, _ = binary.ReadUvarint(packetReader)
		_, _ = binary.ReadUvarint(packetReader)
		_, _ = binary.ReadUvarint(packetReader)
		serverAddrSize, _ := binary.ReadUvarint(packetReader)

		var serverAddressRaw = make([]byte, int(serverAddrSize))
		packetReader.Read(serverAddressRaw)

		route, err := server.GetRoute(string(serverAddressRaw))
		if err != nil {
			log.Error().
				Err(err).
				Msg("No route has been created for the provided address")
			return
		}

		log.Info().
			Str("ProvidedAddress", string(serverAddressRaw)).
			Str("ResolvedAddress", route).
			Msg("Resolved address")

		serverConn, err := net.Dial(
			"tcp",
			fmt.Sprintf("%s:25565", route))

		serverConn.Write(packet)

		err = tcp.Proxy(clientConn, serverConn)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Error while proxying client")
			return
		}

		log.Info().
			Msg("Finished proxying client")
	})

	listener := k8s.NewListener(internal.GetKubernetesClient(), func(serviceMappings []k8s.HostnameToService) {
		log.Debug().
			Interface("ServiceMappings", serviceMappings).
			Msg("HostnameToService received")

		for _, mapping := range serviceMappings {

			if mapping.Delete {
				tcpServer.DeleteRoute(mapping.Hostname)
			} else {
				tcpServer.AddRoute(
					mapping.Hostname,
					fmt.Sprintf("%s.%s.svc.cluster.local", mapping.Service.Name, mapping.Service.Namespace))
			}

		}
	})

	eg, ctx := errgroup.WithContext(context.Background())

	eg.Go(func() error {
		return listener.Run(ctx)
	})

	eg.Go(func() error {
		return tcpServer.Run(ctx)
	})

	if err := eg.Wait(); err != nil {
		log.Fatal().Err(err)
	}
}
