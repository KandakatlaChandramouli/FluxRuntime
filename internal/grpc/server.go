package grpc

import (
	"net"

	ticketv1 "github.com/research/phase1a/internal/gen/ticket/v1"
	"google.golang.org/grpc"
)

// Server wraps the gRPC server with the Phase 1A handler registered.
type Server struct {
	grpc *grpc.Server
	ln   net.Listener
}

// NewServer constructs and binds the gRPC server.
// Interceptors: panic recovery only. No middleware chains.
func NewServer(addr string, handler *Handler) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			PanicRecoveryInterceptor(),
		),
	)

	ticketv1.RegisterTicketReservationServiceServer(srv, handler)

	return &Server{grpc: srv, ln: ln}, nil
}

// Serve begins accepting connections. Blocks until the server stops.
func (s *Server) Serve() error {
	return s.grpc.Serve(s.ln)
}

// Stop performs a graceful shutdown.
func (s *Server) Stop() {
	s.grpc.GracefulStop()
}
