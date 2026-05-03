package transport

import (
	"net"
	"sync"

	transportpb "helios/generated/transport"
	tree "helios/internal/component_tree"
)

type CommClient struct{}

type Sender interface {
	ClientAddress() string
	SendTransportMessage(message *transportpb.TransportMessage)
	SendError(address string, eventName string, code transportpb.EventError_ErrorCode, message string, requestID *string)
}

type TransportService struct {
	mu                   sync.RWMutex
	tree                 *tree.ComponentTree
	subscriptions        map[string]subscription
	subscriptionsByEvent map[eventKey]map[string]struct{}
}

type eventKey struct {
	address   string
	eventName string
}

type subscription struct {
	id     string
	key    eventKey
	owner  string
	sender Sender
}

func NewTransportService(tree *tree.ComponentTree) *TransportService {
	return &TransportService{
		tree:                 tree,
		subscriptions:        make(map[string]subscription),
		subscriptionsByEvent: make(map[eventKey]map[string]struct{}),
	}
}

func (s *TransportService) AttachTransportConnection(address string, conn net.Conn, mustBeRegistered bool) error {
	return s.tree.AttachConnection(address, conn, mustBeRegistered)
}

func (s *TransportService) DetachTransportConnection(address string) {
	s.tree.DetachConnection(address)
	s.removeSubscriptionsForClient(address)
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
	case *transportpb.TransportMessage_EventSubscribe:
		s.handleEventSubscribe(sender, typed.EventSubscribe)
	case *transportpb.TransportMessage_EventUnsubscribe:
		s.handleEventUnsubscribe(sender, typed.EventUnsubscribe)
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
		return
	}
	s.publishToSubscribers(message)
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

func (s *TransportService) handleEventSubscribe(sender Sender, message *transportpb.EventSubscribe) {
	if message == nil {
		return
	}

	subscriptionID := message.GetSubscriptionId()
	if subscriptionID == "" {
		sender.SendError(message.GetAddress(), message.GetEventName(), transportpb.EventError_INVALID_REQUEST, "subscription id is empty", nil)
		return
	}
	if _, _, err := s.tree.GetEvent(message.GetAddress(), message.GetEventName()); err != nil {
		sender.SendError(message.GetAddress(), message.GetEventName(), transportpb.EventError_INVALID_REQUEST, err.Error(), &subscriptionID)
		return
	}

	s.addSubscription(subscription{
		id: subscriptionID,
		key: eventKey{
			address:   message.GetAddress(),
			eventName: message.GetEventName(),
		},
		owner:  sender.ClientAddress(),
		sender: sender,
	})
}

func (s *TransportService) handleEventUnsubscribe(sender Sender, message *transportpb.EventUnsubscribe) {
	if message == nil {
		return
	}

	subscriptionID := message.GetSubscriptionId()
	if subscriptionID == "" {
		sender.SendError("", "", transportpb.EventError_INVALID_REQUEST, "subscription id is empty", nil)
		return
	}
	s.removeSubscription(subscriptionID)
}

func (s *TransportService) addSubscription(sub subscription) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.subscriptions[sub.id]; ok {
		delete(s.subscriptionsByEvent[existing.key], existing.id)
	}

	s.subscriptions[sub.id] = sub
	if s.subscriptionsByEvent[sub.key] == nil {
		s.subscriptionsByEvent[sub.key] = make(map[string]struct{})
	}
	s.subscriptionsByEvent[sub.key][sub.id] = struct{}{}
}

func (s *TransportService) removeSubscription(subscriptionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, ok := s.subscriptions[subscriptionID]
	if !ok {
		return
	}

	delete(s.subscriptions, subscriptionID)
	delete(s.subscriptionsByEvent[sub.key], subscriptionID)
	if len(s.subscriptionsByEvent[sub.key]) == 0 {
		delete(s.subscriptionsByEvent, sub.key)
	}
}

func (s *TransportService) removeSubscriptionsForClient(address string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for subscriptionID, sub := range s.subscriptions {
		if sub.owner != address {
			continue
		}
		delete(s.subscriptions, subscriptionID)
		delete(s.subscriptionsByEvent[sub.key], subscriptionID)
		if len(s.subscriptionsByEvent[sub.key]) == 0 {
			delete(s.subscriptionsByEvent, sub.key)
		}
	}
}

func (s *TransportService) publishToSubscribers(message *transportpb.EventPublish) {
	subscribers := s.subscribersFor(eventKey{
		address:   message.GetAddress(),
		eventName: message.GetEventName(),
	})

	for _, sub := range subscribers {
		sub.sender.SendTransportMessage(&transportpb.TransportMessage{
			Message: &transportpb.TransportMessage_EventPublish{
				EventPublish: &transportpb.EventPublish{
					Address:   message.GetAddress(),
					EventName: message.GetEventName(),
					Event:     message.GetEvent(),
					RequestId: &sub.id,
				},
			},
		})
	}
}

func (s *TransportService) subscribersFor(key eventKey) []subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := s.subscriptionsByEvent[key]
	subscribers := make([]subscription, 0, len(ids))
	for id := range ids {
		subscribers = append(subscribers, s.subscriptions[id])
	}
	return subscribers
}
