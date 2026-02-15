package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/devathh/xvibe/message-gateway/internal/app"
)

func main() {
	app, cleanup, err := app.New()
	if err != nil {
		slog.Error("failed to setup app", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer cleanup()

	go func() {
		if err := app.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server", slog.String("error", err.Error()))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	app.Shutdown(context.Background())
}
