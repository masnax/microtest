package state

import "github.com/masnax/microtest/v2/internal/state"

// State exposes the internal daemon state for use with extended API handlers.
type State = state.State

// Hooks exposes the Hooks struct to be imported by the upstream project.
type Hooks = state.Hooks
