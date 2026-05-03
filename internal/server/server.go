package server

import (
	"context"
	"errors"
	"fmt"
	"net"

	tree "helios/internal/component_tree"
	"helios/internal/logger"
	"helios/internal/transport"
)

type Server struct {
	listener  net.Listener
	transport *transport.TransportService
}

func StartServer(ctx context.Context, addr string, tree *tree.ComponentTree) (*Server, error) {
	if tree == nil {
		return nil, fmt.Errorf("component tree is nil")
	}

	logger.Infow("Starting server", "address", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener:  listener,
		transport: transport.NewTransportService(tree),
	}

	go server.listenForConnections(ctx)

	logger.Info("Server started and listening for connections")

	return server, nil
}

func (s *Server) listenForConnections(ctx context.Context) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) || isPeerOrLocalClose(err) {
				return
			}
			logger.Errorw("Error accepting connection", "error", err)
			continue
		}

		go s.handleConnection(ctx, conn)
	}
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	logger.Infow("New connection", "remote_address", conn.RemoteAddr())

	handler := NewConnectionHandler(ctx, conn, s.transport)
	handler.Handle()
}

func (s *Server) Close() error {
	return s.listener.Close()
}
