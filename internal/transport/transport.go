package transport

import (
	"net"

	transportpb "helios/generated/transport"
	tree "helios/internal/component_tree"
)

type CommClient struct{}

type Sender interface {
	SendTransportMessage(message *transportpb.TransportMessage)
	SendError(address string, eventName string, code transportpb.EventError_ErrorCode, message string, requestID *string)
}

type TransportService struct {
	tree *tree.ComponentTree
}

func NewTransportService(tree *tree.ComponentTree) *TransportService {
	return &TransportService{tree: tree}
}

func (s *TransportService) AttachTransportConnection(address string, conn net.Conn, mustBeRegistered bool) error {
	return s.tree.AttachConnection(address, conn, mustBeRegistered)
}

func (s *TransportService) DetachTransportConnection(address string) {
	s.tree.DetachConnection(address)
}

func (s *TransportService) HandleTransportMessage(sender Sender, message *transportpb.TransportMessage) error {
	if message == nil {
		return nil
	}

	switch typed := message.GetMessage().(type) {
	case *transportpb.TransportMessage_EventPublish:
		s.handleEventPublish(sender, typed.EventPublish)
	case *transportpb.TransportMessage_EventRequest:
		s.handleEventRequest(sender, typed.EventRequest)
	default:
	}
	return nil
}

func (s *TransportService) handleEventPublish(sender Sender, message *transportpb.EventPublish) {
	if message == nil {
		return
	}
	if err := s.tree.SetEvent(message.GetAddress(), message.GetEventName(), message.GetEvent()); err != nil {
		sender.SendError(
			message.GetAddress(),
			message.GetEventName(),
			transportpb.EventError_INVALID_REQUEST,
			err.Error(),
			message.RequestId,
		)
	}
}

func (s *TransportService) handleEventRequest(sender Sender, message *transportpb.EventRequest) {
	if message == nil {
		return
	}

	event, ok, err := s.tree.GetEvent(message.GetAddress(), message.GetEventName())
	if err != nil {
		sender.SendError(
			message.GetAddress(),
			message.GetEventName(),
			transportpb.EventError_INVALID_REQUEST,
			err.Error(),
			message.RequestId,
		)
		return
	}
	if !ok {
		sender.SendError(
			message.GetAddress(),
			message.GetEventName(),
			transportpb.EventError_NOT_FOUND,
			"event not found",
			message.RequestId,
		)
		return
	}

	sender.SendTransportMessage(&transportpb.TransportMessage{
		Message: &transportpb.TransportMessage_EventPublish{
			EventPublish: &transportpb.EventPublish{
				Address:   message.GetAddress(),
				EventName: message.GetEventName(),
				Event:     event,
				RequestId: message.RequestId,
			},
		},
	})
}
