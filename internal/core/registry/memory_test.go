package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister_And_Lookup(t *testing.T) {
	r := NewMemoryRegistry()

	err := r.Register("users", []string{"users.GetProfile", "users.GetFollowing"})
	require.NoError(t, err)

	module, err := r.Lookup("users.GetProfile")
	require.NoError(t, err)
	assert.Equal(t, "users", module)

	module, err = r.Lookup("users.GetFollowing")
	require.NoError(t, err)
	assert.Equal(t, "users", module)
}

func TestLookup_NotFound(t *testing.T) {
	r := NewMemoryRegistry()

	_, err := r.Lookup("unknown.query")
	assert.ErrorIs(t, err, ErrQueryNotFound)
}

func TestRegister_DuplicateQuery(t *testing.T) {
	r := NewMemoryRegistry()

	err := r.Register("users", []string{"shared.query"})
	require.NoError(t, err)

	err = r.Register("posts", []string{"shared.query"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestDeregister(t *testing.T) {
	r := NewMemoryRegistry()

	err := r.Register("users", []string{"users.GetProfile"})
	require.NoError(t, err)

	err = r.Deregister("users")
	require.NoError(t, err)

	_, err = r.Lookup("users.GetProfile")
	assert.ErrorIs(t, err, ErrQueryNotFound)
}

func TestDeregister_Unknown(t *testing.T) {
	r := NewMemoryRegistry()

	err := r.Deregister("unknown")
	assert.NoError(t, err)
}
