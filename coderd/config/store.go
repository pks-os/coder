package config

import (
	"sync"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
)

type Store interface {
	PutOrgSetting(orgID uuid.UUID, name string, value string) error
	GetOrgSetting(orgID uuid.UUID, name string) (string, error)
}

var EntryNotFound = xerrors.New("entry not found")

type fakeStore struct {
	mu    sync.Mutex
	store map[uuid.UUID]map[string]string
}

func newFakeStore() *fakeStore {
	return &fakeStore{store: make(map[uuid.UUID]map[string]string)}
}

func (f *fakeStore) PutOrgSetting(orgID uuid.UUID, name string, value string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	st := f.orgStore(orgID)
	st[name] = value
	return nil
}

func (f *fakeStore) GetOrgSetting(orgID uuid.UUID, name string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	st := f.orgStore(orgID)

	out, ok := st[name]
	if !ok {
		return "", EntryNotFound
	}

	return out, nil
}

// orgStore MUST be called with a lock held.
func (f *fakeStore) orgStore(orgID uuid.UUID) map[string]string {
	m, ok := f.store[orgID]
	if !ok {
		f.store[orgID] = map[string]string{}
		m = f.store[orgID]
	}
	return m
}
