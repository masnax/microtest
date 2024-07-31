// Package types provides shared types and structs.
package types

import "github.com/masnax/microtest/v2/rest/types"

// ExtendedType is an example of an API type usable by MicroCluster but defined by this example project.
type ExtendedType struct {
	Sender  types.AddrPort `json:"sender" yaml:"sender"`
	Message string         `json:"message" yaml:"message"`
}
