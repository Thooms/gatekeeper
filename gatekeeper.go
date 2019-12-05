package gatekeeper

import (
	"context"
	"errors"
)

// Unique identifier for an API client.
type Key string

// Statistics about API usage
type Stats struct {
	// Number of remaining requests for this key
	Remaining int64
	// Total limit for this key
	Limit int64
}

var ErrUnknownKey = errors.New("unknown API key")

// Main interface to interact with a keeper.
// Whenever the key is not found, a keeper should return ErrUnknownKey
type Keeper interface {
	// Check if the key is allowed to make a call
	Allow(context.Context, Key) (bool, Stats, error)

	// Return stats for this key
	Stats(context.Context, Key) (Stats, error)
}
