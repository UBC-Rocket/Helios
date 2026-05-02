package main

import (
	"context"
	"fmt"
	"helios/internal/core"
	"helios/internal/logger"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"
)

func main() {
	root := &cli.Command{
		Name: "helios",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "address",
				Usage:   "address to listen on",
				Value:   "0.0.0.0:5000",
				Aliases: []string{"a"},
			},
			&cli.BoolFlag{
				Name:    "docker",
				Aliases: []string{"d"},
				Usage:   "start configured components in Docker",
			},
			&cli.StringFlag{
				Name:    "docker-socket",
				Aliases: []string{"D"},
				Usage:   "path to the Docker socket",
				Value:   "/var/run/docker.sock",
			},
			&cli.StringFlag{
				Name:  "component-tree",
				Usage: "path to component tree configuration",
			},
			&cli.StringFlag{
				Name:    "runtime-hash",
				Aliases: []string{"rt"},
				Usage:   "runtime identifier hash",
			},
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "logging level: error, warn, info, debug, verbose",
				Value:   "info",
				Aliases: []string{"L"},
			},
		},
		Action: runHelios,
	}

	err := root.Run(context.Background(), os.Args)
	if err != nil {
		logger.Errorw("Error running Helios", "error", err)
		os.Exit(1)
	}
}

func runHelios(ctx context.Context, cmd *cli.Command) error {
	logger.MustInit(cmd.String("log-level"), os.Stderr)
	defer func() { _ = logger.Sync() }()

	logger.Info("Starting Helios")
	logger.Debugw("Arguments", "args", os.Args)

	// Cancel on Ctrl-C (SIGINT) or SIGTERM so work tied to this context stops cleanly.
	shutdownCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	helios := core.Initialize(shutdownCtx, cmd.String("runtime-hash"), cmd.Bool("docker"))
	defer helios.Close()

	// TODO: Introduce other ways of loading the component tree that is not just a file.
	if err := helios.InitializeComponentTree(cmd.String("component-tree")); err != nil {
		logger.Errorw("Failed to initialize component tree", "error", err)
		return fmt.Errorf("Failed to initialize component tree: %w", err)
	}

	if err := helios.InitializeDockerRuntime(cmd.String("docker-socket")); err != nil {
		logger.Errorw("Failed to initialize Docker runtime", "error", err)
		return fmt.Errorf("Failed to initialize Docker runtime: %w", err)
	}

	if err := helios.InitializeServer(cmd.String("address")); err != nil {
		logger.Errorw("Failed to initialize server", "error", err)
		return fmt.Errorf("Failed to initialize server: %w", err)
	}

	logger.Info("Running...")
	<-shutdownCtx.Done()
	logger.Info("shutting down")
	return nil
}
