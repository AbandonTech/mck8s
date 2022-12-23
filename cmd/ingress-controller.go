package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	"minecraftk8s/internal"
	"minecraftk8s/pkg/k8s"
	"minecraftk8s/pkg/tcp"
	"net"
	"os"
)

const VERSION = "v0.0.1"

func main() {
	// Overwrite version flag to be capital 'V'
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print the version",
	}

	app := &cli.App{
		Name:  "mck8s-ingress-controller",
		Usage: "route minecraft traffic to k8s annotated services",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "verbose", Usage: "Set logging to debug level", Aliases: []string{"v"}},
			&cli.PathFlag{Name: "kubeconfig", Usage: "Path to kubeconfig"},
			&cli.IntFlag{Name: "port", Usage: "Port to accept connections via", Value: 25565},
		},
		Action: cliMain,
		Before: func(ctx *cli.Context) error {
			internal.ConfigureLogging(ctx.Bool("verbose"))
			return nil
		},
		Version: VERSION,
		Authors: []*cli.Author{{
			Name:  "GDWR",
			Email: "gregory.dwr@gmail.com",
		}},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().
			Err(err).
			Msg("Error while running cli")
	}
}

func cliMain(ctx *cli.Context) error {
	log.Info().
		Msg("Initializing mck8s-ingress-controller")

	tcpServer := tcp.NewTcpServer(ctx.Int("port"), func(server *tcp.TcpServer, clientConn net.Conn) {
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
		serverAddress := string(serverAddressRaw)

		route, err := server.GetRoute(serverAddress)
		if err != nil {
			log.Error().
				Err(err).
				Str("RequestedAddress", serverAddress).
				Msg("No route has been created for the provided address")
			return
		}

		log.Info().
			Str("ProvidedAddress", string(serverAddressRaw)).
			Str("ResolvedAddress", route).
			Msg("Resolved address")

		serverConn, err := net.Dial("tcp", fmt.Sprintf("%s:25565", route))
		if err != nil {
			log.Error().
				Err(err).
				Msg("Unable to connect to server")
			return

		}

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

	k8sClient := internal.GetKubernetesClient(internal.GetKubernetesConfig(ctx.Path("kubeconfig")))
	listener := k8s.NewListener(k8sClient, func(serviceMappings []k8s.ServiceMapping) {
		for _, mapping := range serviceMappings {

			if mapping.Delete {
				tcpServer.DeleteRoute(mapping.Hostname)
			} else {
				tcpServer.AddRoute(
					mapping.Hostname,
					mapping.Service.Spec.ClusterIP)
				//fmt.Sprintf("%s.%s.svc.cluster.local", mapping.Service.Name, mapping.Service.Namespace))
			}

		}
	})

	eg, errCtx := errgroup.WithContext(context.Background())

	eg.Go(func() error {
		return listener.Run(errCtx)
	})

	eg.Go(func() error {
		return tcpServer.Run(errCtx)
	})

	if err := eg.Wait(); err != nil {
		log.Fatal().Err(err)
	}
	return nil
}
