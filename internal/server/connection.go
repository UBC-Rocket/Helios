package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"

	"helios/internal/logger"
	"helios/internal/transport"

	transportpb "helios/generated/transport"
)

const MAX_MAILBOX_SIZE = 128

type ClientInfo struct {
	Version          int
	Address          string
	MustBeRegistered bool
}

type ConnectionHandler struct {
	ctx       context.Context
	conn      net.Conn
	transport *transport.TransportService
	client    ClientInfo
	mailbox   chan *transportpb.TransportMessage
}

func NewConnectionHandler(ctx context.Context, conn net.Conn, transportService *transport.TransportService) *ConnectionHandler {
	return &ConnectionHandler{
		ctx:       ctx,
		conn:      conn,
		transport: transportService,
		mailbox:   make(chan *transportpb.TransportMessage, MAX_MAILBOX_SIZE),
	}
}

func (h *ConnectionHandler) Handle() {
	defer h.Close()

	clientInfo, err := h.performHandshake()
	if err != nil {
		if h.ctx.Err() != nil {
			return
		}

		if isPeerOrLocalClose(err) {
			logger.Debugw("Connection closed", "remote_address", h.conn.RemoteAddr(), "reason", err)
			return
		}

		logger.Errorw("Error reading from connection", "remote_address", h.conn.RemoteAddr(), "error", err)
		return
	}
	h.client = clientInfo

	logger.Debugw("Handshake successful", "remote_address", h.conn.RemoteAddr(), "client_info", h.client)

	go h.readLoop()
	go h.writeLoop()

	<-h.ctx.Done()
	logger.Infow("Closing connection", "remote_address", h.conn.RemoteAddr())
}

func (h *ConnectionHandler) performHandshake() (ClientInfo, error) {
	transportMessage, err := ReadMessage(h.conn)
	if err != nil {
		return ClientInfo{}, err
	}

	handshakeRequest := transportMessage.GetHandshakeRequest()
	if handshakeRequest == nil {
		return ClientInfo{}, fmt.Errorf("expected handshake request, got: %v", transportMessage)
	}

	logger.Infow("Handshake request received", "remote_address", h.conn.RemoteAddr(), "message", handshakeRequest)

	clientVersion := int(handshakeRequest.GetVersion())
	if clientVersion != PROTOCOL_VERSION {
		return ClientInfo{}, fmt.Errorf("unsupported protocol version from client: %d", handshakeRequest.GetVersion())
	}

	clientAddress := handshakeRequest.GetClientAddress()
	mustBeRegistered := handshakeRequest.GetMustBeRegistered()
	if clientAddress == "" {
		return ClientInfo{}, fmt.Errorf("client address is empty")
	}

	if err := h.transport.AttachTransportConnection(clientAddress, h.conn, mustBeRegistered); err != nil {
		return ClientInfo{}, err
	}
	clientInfo := ClientInfo{
		Version:          clientVersion,
		Address:          clientAddress,
		MustBeRegistered: mustBeRegistered,
	}

	handshakeResponse := &transportpb.TransportMessage{
		Message: &transportpb.TransportMessage_HandshakeResponse{
			HandshakeResponse: &transportpb.HandshakeResponse{
				Version: PROTOCOL_VERSION,
			},
		},
	}

	err = SendMessage(h.conn, handshakeResponse)
	if err != nil {
		return ClientInfo{}, err
	}

	logger.Infow("Handshake response sent", "remote_address", h.conn.RemoteAddr(), "message", handshakeResponse)

	return clientInfo, nil
}

func (h *ConnectionHandler) readLoop() {
	for {
		transportMessage, err := ReadMessage(h.conn)
		if err != nil {
			if h.ctx.Err() != nil {
				return
			}

			if isPeerOrLocalClose(err) {
				logger.Debugw("Connection closed", "remote_address", h.conn.RemoteAddr(), "reason", err)
				return
			}

			logger.Errorw("Error reading from connection", "remote_address", h.conn.RemoteAddr(), "error", err)
			return
		}

		logger.Debugw("Message received from client", "remote_address", h.conn.RemoteAddr(), "message", transportMessage)
		if err := h.transport.HandleTransportMessage(h, transportMessage); err != nil {
			logger.Errorw("Error handling transport message", "remote_address", h.conn.RemoteAddr(), "error", err)
		}
	}
}

func (h *ConnectionHandler) writeLoop() {
	for {
		select {
		case message := <-h.mailbox:
			if err := SendMessage(h.conn, message); err != nil {
				if h.ctx.Err() != nil {
					return
				}
				if isPeerOrLocalClose(err) {
					logger.Debugw("Connection closed", "remote_address", h.conn.RemoteAddr(), "reason", err)
					return
				}
				logger.Errorw("Error writing to connection", "remote_address", h.conn.RemoteAddr(), "error", err)
			}
		case <-h.ctx.Done():
			return
		}
	}
}

func (h *ConnectionHandler) Close() {
	h.transport.DetachTransportConnection(h.client.Address)
	h.conn.Close()
}

func (h *ConnectionHandler) ClientAddress() string {
	return h.client.Address
}

func (h *ConnectionHandler) SendTransportMessage(message *transportpb.TransportMessage) {
	if message == nil {
		return
	}
	select {
	case h.mailbox <- message:
	case <-h.ctx.Done():
	}
}

func (h *ConnectionHandler) SendError(address string, eventName string, code transportpb.EventError_ErrorCode, message string, requestID *string) {
	h.SendTransportMessage(&transportpb.TransportMessage{
		Message: &transportpb.TransportMessage_EventError{
			EventError: &transportpb.EventError{
				Address:   address,
				EventName: eventName,
				ErrorCode: code,
				Message:   message,
				RequestId: requestID,
			},
		},
	})
}

// isPeerOrLocalClose reports errors that usually mean the TCP session ended
// normally or abruptly on the other side, or we closed the socket.
func isPeerOrLocalClose(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return isPeerOrLocalClose(opErr.Err)
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		switch errno {
		case syscall.ECONNRESET, syscall.EPIPE, syscall.ECONNABORTED, syscall.ENOTCONN:
			return true
		}
	}
	return false
}
