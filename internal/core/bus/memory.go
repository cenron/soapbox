package bus

import (
	"fmt"
	"log/slog"
	"sync"
)

type memoryBus struct {
	mu      sync.RWMutex
	subs    map[string][]func(event any)
	queries map[string]func(req any) (any, error)
	logger  *slog.Logger
}

func NewMemoryBus(logger *slog.Logger) Bus {
	return &memoryBus{
		subs:    make(map[string][]func(event any)),
		queries: make(map[string]func(req any) (any, error)),
		logger:  logger,
	}
}

func (b *memoryBus) Publish(topic string, event any) error {
	b.mu.RLock()
	original := b.subs[topic]
	handlers := make([]func(event any), len(original))
	copy(handlers, original)
	b.mu.RUnlock()

	for _, h := range handlers {
		go func(handler func(event any)) {
			defer func() {
				if r := recover(); r != nil {
					b.logger.Error("bus handler panicked", "topic", topic, "panic", r)
				}
			}()
			handler(event)
		}(h)
	}

	return nil
}

func (b *memoryBus) Subscribe(topic string, handler func(event any)) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subs[topic] = append(b.subs[topic], handler)
	return nil
}

func (b *memoryBus) RegisterQuery(name string, handler func(req any) (any, error)) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.queries[name]; exists {
		panic(fmt.Sprintf("bus: query %q already registered", name))
	}

	b.queries[name] = handler
}

func (b *memoryBus) Query(name string, req any) (any, error) {
	b.mu.RLock()
	handler, exists := b.queries[name]
	b.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("bus: query %q not registered", name)
	}

	return handler(req)
}
