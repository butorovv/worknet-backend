package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/butorovv/worknet-backend/internal/app"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := app.New(ctx)
	if err != nil {
		return err
	}

	errc := make(chan error, 1)
	go func() {
		errc <- application.Start()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), application.ShutdownTimeout())
		defer cancel()
		return application.Shutdown(shutdownCtx)
	case err := <-errc:
		return err
	}
}
