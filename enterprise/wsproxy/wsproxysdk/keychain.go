package wsproxysdk

import (
	"context"
	"sync"
	"time"

	"golang.org/x/xerrors"

	"github.com/coder/quartz"
)

type WSProxyKeychain struct {
	client *Client
	clock  quartz.Clock
	keysMu sync.Mutex
	keys   map[int32]SigningKey
}

func (k *WSProxyKeychain) Latest(ctx context.Context) (SigningKey, error) {
	k.keysMu.Lock()
	defer k.keysMu.Unlock()

	latest := findLatestKey(k.keys, k.clock.Now().UTC())
	if latest.StartsAt.IsZero() {
		var err error
		k.keys, err = k.fetch(ctx)
		if err != nil {
			return SigningKey{}, xerrors.Errorf("fetch: %w", err)
		}
	}

	latest = findLatestKey(k.keys, k.clock.Now().UTC())
	if latest.StartsAt.IsZero() {
		return SigningKey{}, xerrors.Errorf("no keys found")
	}

	return latest, nil
}

func (k *WSProxyKeychain) Version(ctx context.Context, sequence int32) (SigningKey, error) {
	k.keysMu.Lock()
	defer k.keysMu.Unlock()

	key, ok := k.keys[sequence]
	if !ok {
		var err error
		k.keys, err = k.fetch(ctx)
		if err != nil {
			return SigningKey{}, xerrors.Errorf("fetch: %w", err)
		}
		key, ok = k.keys[sequence]
	}

	now := k.clock.Now().UTC()
	deleted := !key.DeletesAt.IsZero() && !now.Before(key.DeletesAt)
	if !ok || deleted {
		return SigningKey{}, xerrors.Errorf("key not found")
	}

	return key, nil
}

func (k *WSProxyKeychain) fetch(ctx context.Context) (map[int32]SigningKey, error) {
	keys, err := k.client.SecurityKeys(ctx)
	if err != nil {
		return nil, xerrors.Errorf("get security keys: %w", err)
	}

	kmap := toKeyMap(keys.SigningKeys)
	return kmap, nil
}

func toKeyMap(keys []SigningKey) map[int32]SigningKey {
	m := make(map[int32]SigningKey)
	for _, key := range keys {
		m[key.Sequence] = key
	}
	return m
}

func findLatestKey(keys map[int32]SigningKey, now time.Time) SigningKey {
	var latest SigningKey
	for sequence, key := range keys {
		if sequence > latest.Sequence && key.IsActive(now) {
			latest = key
		}
	}
	return latest
}
