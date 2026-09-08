package internalhttp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
	host       string
	port       string
	logger     Logger
	app        Application
}

type Logger interface {
	Info(msg string)
}

type Application interface{}

func NewServer(logger Logger, app Application, host, port string) *Server {
	return &Server{
		host:   host,
		port:   port,
		logger: logger,
		app:    app,
	}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.hello)

	s.httpServer = &http.Server{
		Addr:              net.JoinHostPort(s.host, s.port),
		Handler:           loggingMiddleware(mux, s.logger),
		ReadHeaderTimeout: 10 * time.Second, // защита от Slowloris (G112)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return s.Stop(ctx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) hello(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("hello-world"))
}
