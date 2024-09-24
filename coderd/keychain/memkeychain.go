package keychain

import (
	"context"
	"slices"

	"github.com/coder/coder/v2/enterprise/wsproxy/wsproxysdk"
)

type WSProxyKeychain struct {
	keys []wsproxysdk.SigningKey
}

func NewMemKeychain(keys []wsproxysdk.SigningKey) *WSProxyKeychain {
	slices.SortFunc(keys, func(a, b wsproxysdk.SigningKey) int {
		return int(a.Sequence - b.Sequence)
	})
	return &WSProxyKeychain{keys: keys}
}

func (k *WSProxyKeychain) Latest(context.Context) ([]byte, error) {
	if len(k.keys) == 0 {
		return nil, ErrKeyNotFound
	}

	key := k.keys[len(k.keys)-1]
	return []byte(key.Secret), nil
}

func (k *WSProxyKeychain) KeyVersion(_ context.Context, sequence int32) ([]byte, error) {
	for i := len(k.keys) - 1; i >= 0; i-- {
		key := k.keys[i]
		if key.Sequence != sequence {
			continue
		}
		return []byte(key.Secret), nil
	}
	return nil, ErrKeyNotFound
}
