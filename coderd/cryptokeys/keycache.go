package cryptokeys

import (
	"context"

	"golang.org/x/xerrors"
)

var ErrKeyNotFound = xerrors.New("key not found")

var ErrKeyInvalid = xerrors.New("key is invalid for use")

// Keycache provides an abstraction for fetching signing keys.
type Keycache interface {
	Latest(ctx context.Context) ([]byte, error)
	Version(ctx context.Context, sequence int32) ([]byte, error)
}
