package config

import (
	"sync"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
)

type Store interface {
	GetOrgSetting(orgID uuid.UUID, name string) (string, error)
	PutOrgSetting(orgID uuid.UUID, name string, value string) error
	DeleteOrgSetting(orgID uuid.UUID, name string) error
}

var EntryNotFound = xerrors.New("entry not found")

type FakeStore struct {
	mu    sync.Mutex
	store map[uuid.UUID]map[string]string
}

func NewFakeStore() *FakeStore {
	return &FakeStore{store: make(map[uuid.UUID]map[string]string)}
}

func (f *FakeStore) PutOrgSetting(orgID uuid.UUID, name string, value string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	st := f.orgStore(orgID)
	st[name] = value
	return nil
}

func (f *FakeStore) GetOrgSetting(orgID uuid.UUID, name string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	st := f.orgStore(orgID)

	out, ok := st[name]
	if !ok {
		return "", EntryNotFound
	}

	return out, nil
}

func (f *FakeStore) DeleteOrgSetting(orgID uuid.UUID, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	st := f.orgStore(orgID)
	delete(st, name)
	return nil
}

// orgStore MUST be called with a lock held.
func (f *FakeStore) orgStore(orgID uuid.UUID) map[string]string {
	m, ok := f.store[orgID]
	if !ok {
		f.store[orgID] = map[string]string{}
		m = f.store[orgID]
	}
	return m
}
