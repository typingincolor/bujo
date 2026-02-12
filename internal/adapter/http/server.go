package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/typingincolor/bujo/internal/service"
)

const DefaultPort = 8743

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
	if s.port < 0 {
		return "127.0.0.1:0"
	}
	port := s.port
	if port == 0 {
		port = DefaultPort
	}
	return fmt.Sprintf("127.0.0.1:%d", port)
}

func (s *Server) Start() (string, error) {
	addr := s.listenAddr()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = ln

	go func() { _ = s.httpServer.Serve(ln) }()

	return ln.Addr().String(), nil
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
