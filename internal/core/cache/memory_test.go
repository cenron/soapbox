package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet_And_Get(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	err := c.Set(ctx, "key", "value", time.Minute)
	require.NoError(t, err)

	var result string
	err = c.Get(ctx, "key", &result)
	require.NoError(t, err)
	assert.Equal(t, "value", result)
}

func TestGet_Miss(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	var result string
	err := c.Get(ctx, "missing", &result)
	assert.ErrorIs(t, err, ErrCacheMiss)
}

func TestGet_Expired(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	err := c.Set(ctx, "key", "value", time.Millisecond)
	require.NoError(t, err)

	time.Sleep(5 * time.Millisecond)

	var result string
	err = c.Get(ctx, "key", &result)
	assert.ErrorIs(t, err, ErrCacheMiss)
}

func TestDelete(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	err := c.Set(ctx, "key", "value", time.Minute)
	require.NoError(t, err)

	err = c.Delete(ctx, "key")
	require.NoError(t, err)

	var result string
	err = c.Get(ctx, "key", &result)
	assert.ErrorIs(t, err, ErrCacheMiss)
}

func TestSet_StructValue(t *testing.T) {
	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	c := NewMemoryCache()
	ctx := context.Background()

	err := c.Set(ctx, "user:1", user{Name: "Alice", Age: 30}, time.Minute)
	require.NoError(t, err)

	var result user
	err = c.Get(ctx, "user:1", &result)
	require.NoError(t, err)
	assert.Equal(t, "Alice", result.Name)
	assert.Equal(t, 30, result.Age)
}
