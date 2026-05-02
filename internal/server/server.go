package server

import (
	"context"
	"errors"
	"net"

	"helios/internal/logger"
)

type Server struct {
	listener net.Listener
	attacher ComponentAttacher
}

type ComponentAttacher interface {
	AttachConnection(address string, conn net.Conn, mustBeRegistered bool) error
}

func StartServer(ctx context.Context, addr string, attacher ComponentAttacher) (*Server, error) {
	logger.Infow("Starting server", "address", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		attacher: attacher,
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

	handler := NewConnectionHandler(ctx, conn, s.attacher)
	handler.Handle()
}

func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}
