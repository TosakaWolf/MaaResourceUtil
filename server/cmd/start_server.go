package main

import (
	"context"
	"errors"
	"golang.org/x/net/http2"
	"maaResourceUtil/server/internal/config"
	"maaResourceUtil/server/internal/server"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func main() {
	//go server_turn.Start()
	e, c := server.New()
	// Start server
	go func() {
		if err := e.StartH2CServer(":"+strconv.Itoa(config.Config.Port), &http2.Server{}); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal("shutting down the server")
		}
	}()
	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	if c != nil {
		c.Close()
	}
}
