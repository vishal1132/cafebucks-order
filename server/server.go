package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/vishal1132/cafebucks/config"
	"github.com/vishal1132/cafebucks/eventbus"
)

// Server is the http server struct
type server struct {
	Mux      *mux.Router
	Logger   zerolog.Logger
	EventBus *eventbus.EB
}

func runserver(cfg config.C, logger zerolog.Logger) error {
	logger = logger.With().Str("context", "order service").Logger()
	// set up signal caching
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT)

	logger.Info().
		Str("env", string(cfg.Env)).
		Str("log_level", cfg.LogLevel.String())

	_, cancel := context.WithCancel(context.Background())

	defer cancel()

	// initialize eventbus
	ebcfg := eventbus.Config{
		Logger: &logger,
		Topic:  "test",
	}
	eb, err := eventbus.New(ebcfg)
	if err != nil {
		logger.Error().Err(err).Msg("Could not create an instance of eventbus")
		return err
	}

	// Create an event reader in a concurrent go routine
	/*
		go func() {
			msg, err := eb.ReadEvents(ctx)
			if err != nil {
				logger.With().Str("event read", "error").Err(err)
			}
			logger.Info().Str("Value", string(msg.Value)).Str("Key", string(msg.Key)).Msg("event received")
		}()
	*/
	go func() {
		sig := <-signalCh

		cancel()

		logger.Info().
			Str("signal", sig.String()).
			Msg("shutting down http server gracefully")
	}()

	s := server{
		Mux:      mux.NewRouter(),
		Logger:   logger,
		EventBus: eb,
	}

	s.registerHandlers()

	httpSrvr := &http.Server{
		Handler:     s.Mux,
		ReadTimeout: 20 * time.Second,
		IdleTimeout: 60 * time.Second,
	}

	// creating a tcp listener
	socketAddr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
	logger.Info().
		Str("addr", socketAddr).
		Msg("binding to TCP socket")

	// set up the network socket
	listener, err := net.Listen("tcp", socketAddr)
	if err != nil {
		return fmt.Errorf("failed to open HTTP socket: %w", err)
	}

	// signal handling / graceful shutdown goroutine
	serveStop, serverShutdown := make(chan struct{}), make(chan struct{})
	var serveErr, shutdownErr error

	// HTTP server parent goroutine
	go func() {
		defer close(serveStop)
		serveErr = httpSrvr.Serve(listener)
	}()

	// signal handling / graceful shutdown goroutine
	go func() {
		defer close(serverShutdown)
		sig := <-signalCh

		logger.Info().
			Str("signal", sig.String()).
			Msg("shutting HTTP server down gracefully")

		cctx, ccancel := context.WithTimeout(context.Background(), 25*time.Second)

		defer ccancel()
		defer cancel()

		if shutdownErr = httpSrvr.Shutdown(cctx); shutdownErr != nil {
			logger.Error().
				Err(shutdownErr).
				Msg("failed to gracefully shut down HTTP server")
		}
	}()

	// wait for it to die
	<-serverShutdown
	<-serveStop

	// log errors for informational purposes
	logger.Info().
		AnErr("serve_err", serveErr).
		AnErr("shutdown_err", shutdownErr).
		Msg("server shut down")
	return nil
}
