package main

import (
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"minecraftk8s/internal"
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
		Name:  "mck8s-cli",
		Usage: "configure mck8s deployments via commandline",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "verbose", Usage: "Set logging to debug level", Aliases: []string{"v"}},
			&cli.PathFlag{Name: "kubeconfig", Usage: "Path to kubeconfig"},
		},
		Action: func(context *cli.Context) error {
			return nil
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
