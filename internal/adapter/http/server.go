package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/typingincolor/bujo/internal/service"
)

const (
	DefaultPort   = 8743
	EphemeralPort = -1
)

type Server struct {
	httpServer *http.Server
	listener   net.Listener
	port       int
}

func NewServer(bujo *service.BujoService, port int) *Server {
	handler := NewHandler(bujo)

	return &Server{
		httpServer: &http.Server{
			Handler: handler.Routes(),
		},
		port: port,
	}
}

func (s *Server) listenAddr() string {
	if s.port == EphemeralPort {
		return "127.0.0.1:0"
	}
	if s.port > 0 {
		return fmt.Sprintf("127.0.0.1:%d", s.port)
	}
	return fmt.Sprintf("127.0.0.1:%d", DefaultPort)
}

func (s *Server) Start() (string, error) {
	addr := s.listenAddr()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = ln

	go func() {
		if err := s.httpServer.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		}
	}()

	return ln.Addr().String(), nil
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
