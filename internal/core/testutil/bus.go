package testutil

import (
	"fmt"
	"sync"
)

type PublishedEvent struct {
	Topic string
	Event any
}

type MockBus struct {
	mu             sync.Mutex
	published      []PublishedEvent
	queryResponses map[string]func(req any) (any, error)
	subscribers    map[string][]func(event any)
}

func NewMockBus() *MockBus {
	return &MockBus{
		queryResponses: make(map[string]func(req any) (any, error)),
		subscribers:    make(map[string][]func(event any)),
	}
}

func (b *MockBus) Publish(topic string, event any) error {
	b.mu.Lock()
	b.published = append(b.published, PublishedEvent{Topic: topic, Event: event})
	original := b.subscribers[topic]
	handlers := make([]func(event any), len(original))
	copy(handlers, original)
	b.mu.Unlock()

	for _, h := range handlers {
		h(event)
	}

	return nil
}

func (b *MockBus) Subscribe(topic string, handler func(event any)) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[topic] = append(b.subscribers[topic], handler)
	return nil
}

func (b *MockBus) RegisterQuery(name string, handler func(req any) (any, error)) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.queryResponses[name] = handler
	return nil
}

func (b *MockBus) Query(name string, req any) (any, error) {
	b.mu.Lock()
	handler, exists := b.queryResponses[name]
	b.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("bus: query %q not registered", name)
	}

	return handler(req)
}

func (b *MockBus) Published() []PublishedEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	result := make([]PublishedEvent, len(b.published))
	copy(result, b.published)
	return result
}

func (b *MockBus) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.published = nil
}
