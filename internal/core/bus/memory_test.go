package bus

import (
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestBus() Bus {
	return NewMemoryBus(slog.Default())
}

func TestPublish_DispatchesToSubscribers(t *testing.T) {
	b := newTestBus()

	var mu sync.Mutex
	var received []string

	err := b.Subscribe("test.topic", func(event any) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, event.(string))
	})
	require.NoError(t, err)

	err = b.Publish("test.topic", "hello")
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []string{"hello"}, received)
}

func TestPublish_MultipleSubscribers(t *testing.T) {
	b := newTestBus()

	var mu sync.Mutex
	count := 0

	for range 3 {
		err := b.Subscribe("test.topic", func(_ any) {
			mu.Lock()
			defer mu.Unlock()
			count++
		})
		require.NoError(t, err)
	}

	err := b.Publish("test.topic", "event")
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 3, count)
}

func TestPublish_NoSubscribers(t *testing.T) {
	b := newTestBus()

	err := b.Publish("no.subscribers", "event")
	assert.NoError(t, err)
}

func TestRegisterQuery_And_Query(t *testing.T) {
	b := newTestBus()

	b.RegisterQuery("test.query", func(req any) (any, error) {
		return "response:" + req.(string), nil
	})

	result, err := b.Query("test.query", "input")
	require.NoError(t, err)
	assert.Equal(t, "response:input", result)
}

func TestQuery_Unregistered(t *testing.T) {
	b := newTestBus()

	_, err := b.Query("unknown.query", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")
}

func TestRegisterQuery_DuplicatePanics(t *testing.T) {
	b := newTestBus()

	b.RegisterQuery("test.query", func(_ any) (any, error) { return nil, nil })

	assert.Panics(t, func() {
		b.RegisterQuery("test.query", func(_ any) (any, error) { return nil, nil })
	})
}
