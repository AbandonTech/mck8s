package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"minecraftk8s/internal"
	"minecraftk8s/internal/api/handlers"
	"minecraftk8s/internal/api/middleware"
	"net/http"
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
		Name:  "mck8s-api",
		Usage: "create, add and/or remove Minecraft Server deployments via an api",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "verbose", Usage: "Set logging to debug level", Aliases: []string{"v"}},
			&cli.PathFlag{Name: "kubeconfig", Usage: "Path to kubeconfig"},
			&cli.StringFlag{Name: "host", Usage: "Address to bind http server to", Value: "127.0.0.1"},
			&cli.IntFlag{Name: "port", Usage: "Port to accept connections via", Value: 8080},
		},
		Action: func(c *cli.Context) error {
			port := c.Int("port")

			mux := http.NewServeMux()
			mux.Handle("/", middleware.LoggingMiddleware(handlers.IndexHandler()))

			log.Info().
				Str("Url", fmt.Sprintf("http://%s:%d", c.String("host"), port)).
				Msg("Starting server")
			return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
		},
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
