package config

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/coder/serpent"
	"github.com/google/uuid"
	"github.com/spf13/pflag"
	"golang.org/x/xerrors"
)

// Option is an alias of pflag.Value; this is the type used internally by serpent.Option, but
// this may change in the future. We also want to hide this implementation detail from consumers of this
// API since it is irrelevant.
type Option pflag.Value

// Manager coordinates storing and retrieve org-level settings against the Store, and implements a look-aside cache.
type Manager struct {
	Store
	options serpent.OptionSet

	mu    sync.RWMutex // Protects cache.
	cache map[string]string
}

func NewManager(options serpent.OptionSet) *Manager {
	if len(options) == 0 {
		panic("developer error: options must not be empty")
	}

	return &Manager{
		options: options,
		cache:   make(map[string]string),
		Store:   newFakeStore(),
	}
}

// lookupByField finds an associated *serpent.Option for a given config field.
// TODO: use a more efficient search than O(N).
func (m *Manager) lookupByField(field Option) *serpent.Option {
	for _, opt := range m.options {
		if opt.Value == field {
			return &opt
		}
	}

	panic("developer error: given field has no associated option defined in codersdk/deployment.go")
}

// ResolveForOrgByOption is not a method on the Manager struct because Go does not support generic methods: https://github.com/golang/go/issues/49085
// We require generics here to enforce and simplify type-safety and to return an instance of the given src type T
// regardless of whether there exists an override for this field or not.
//
// field MUST be a pointer to a config field defined in `func (c *DeploymentValues) Options() serpent.OptionSet`.
//
// The field is used to lookup a *serpent.Option which contains identifying information about the field. We use the Env
// value as the key to index this setting against.
func ResolveForOrgByOption[T Option](m *Manager, orgID uuid.UUID, field T) (T, error) {
	var zero T

	// Find the option definition for the given field.
	opt := m.lookupByField(field)

	// Check if an override exists.
	val, err := m.GetOrgSetting(orgID, opt.Env)
	if err != nil {
		// No override found, return source value.
		if errors.Is(err, EntryNotFound) {
			return field, nil
		}
		return zero, err
	}

	// Override found, construct new instance of T.
	newInst := reflect.New(reflect.TypeOf(field).Elem()).Interface().(T)
	// val is a string, and Option.Set accepts a string to unmarshal into a complex type.
	if err = newInst.Set(val); err != nil {
		return zero, xerrors.Errorf("construct instance of %T: %w", field, err)
	}

	return newInst, nil
}

// ResolveForOrgByName caters for all settings which do not have deployment-level counterparts.
// This could be a method on Manager but is not for consistency with ResolveForOrgByOption.
func ResolveForOrgByName(m *Manager, orgID uuid.UUID, name string) (string, error) {
	// Check if an override exists.
	val, err := m.GetOrgSetting(orgID, name)
	if err != nil {
		return "", err
	}

	return val, nil
}

// AddOrgSettingOverride adds an org-level override for a given setting, using the Env value of the setting as the key.
func (m *Manager) AddOrgSettingOverride(orgID uuid.UUID, src Option, value string) error {
	opt := m.lookupByField(src)
	if opt.Env == "" {
		// TODO: add validation/linter to prevent this
		return xerrors.Errorf("opt has no env set: %q", opt.Flag)
	}

	return m.PutOrgSetting(orgID, opt.Env, value)
}

// AddOrgSettingByName adds an org-level setting which do not have deployment-level counterpart.
func (m *Manager) AddOrgSettingByName(orgID uuid.UUID, name string, value string) error {
	return m.PutOrgSetting(orgID, name, value)
}

// PutOrgSetting adds or updates an org-level setting.
// This method intercepts a call to the store to provide look-aside caching.
func (m *Manager) PutOrgSetting(orgID uuid.UUID, name string, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.Store.PutOrgSetting(orgID, name, value); err != nil {
		return err
	}

	m.cache[m.cacheKey(orgID, name)] = value
	return nil
}

// GetOrgSetting retrieves an org-level setting, should one exist.
// If one does not exist, the sentinal error EntryNotFound is returned.
// This method intercepts a call to the store to provide look-aside caching.
func (m *Manager) GetOrgSetting(orgID uuid.UUID, name string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	val, ok := m.cache[m.cacheKey(orgID, name)]
	if ok {
		return val, nil
	}

	return m.Store.GetOrgSetting(orgID, name)
}

func (m *Manager) cacheKey(orgID uuid.UUID, name string) string {
	return fmt.Sprintf("%s:%s", orgID.String(), name)
}

// Flush flushes the look-aside store cache.
func (m *Manager) Flush() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	clear(m.cache)

	return nil
}
