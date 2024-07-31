package client

import (
	"context"

	"github.com/masnax/microtest/v2/internal/rest/types"
)

// ControlDaemon posts control data to the daemon.
func (c *Client) ControlDaemon(ctx context.Context, args types.Control) error {
	return c.QueryStruct(ctx, "POST", types.ControlEndpoint, nil, args, nil)
}
