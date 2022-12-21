package main

import (
	"context"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"minecraftk8s/internal"
	"minecraftk8s/pkg/k8s"
	"minecraftk8s/pkg/minecraft"
	"minecraftk8s/pkg/tcp_server"
	"net"
)

func main() {
	internal.ConfigureLogging()
	log.Info().Msg("Starting mck8s-ingress-controller")

	listener := k8s.NewListener(internal.GetKubernetesClient(), func(payload *k8s.Payload) {
		log.Info().
			Interface("payload", payload).
			Msg("Payload received")
	})

	tcpServer := tcp_server.NewTcpServer(25565, func(conn net.Conn) {
		remote := conn.RemoteAddr()

		log.Info().
			Str("Client", remote.String()).
			Msg("Connection received")

		reader := minecraft.NewProtocolReader(conn)

		packet, err := reader.Read()
		if err != nil {
			log.Error().
				Err(err).
				Msg("Couldn't read from TCP stream")
			return
		}

		log.Info().
			Interface("Packet", packet).
			Msg("Packet received")
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
