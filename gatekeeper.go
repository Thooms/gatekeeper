package gatekeeper

import (
	"context"
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

// Main interface to interact with a keeper.
type Keeper interface {
	// Check if the key is allowed to make a call
	Allow(context.Context, Key) (bool, Stats, error)

	// Return stats for this key
	Stats(context.Context, Key) (Stats, error)
}
