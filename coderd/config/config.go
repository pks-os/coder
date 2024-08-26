package config

import (
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/coder/serpent"
	"github.com/google/uuid"
	"github.com/spf13/pflag"
	"golang.org/x/xerrors"
)

type Value pflag.Value

type Initializer interface {
	Initialize(opt *serpent.Option)
}

type OrgValuer[T Value] interface {
	Value
	GlobalValue() T
	OrgValue(store Store, orgID uuid.UUID) (T, error)
	Coalesce(store Store, orgID uuid.UUID) T
}

// OrgScoped wraps a Value to provide extra functionality for org-level settings.
type OrgScoped[T Value] struct {
	initialized atomic.Bool

	val T
	opt *serpent.Option
}

// Initialize is intended to be called when the *serpent.Option which describes this value is created, and will run
// at most once.
func (o *OrgScoped[T]) Initialize(opt *serpent.Option) {
	if !o.initialized.CompareAndSwap(false, true) {
		return
	}

	if opt == nil {
		panic("developer error: empty option")
	}

	if opt.Env == "" {
		panic(fmt.Sprintf("developer error: %q has no Env value", opt.Name))
	}

	// Instantiate a new instance of type T.
	o.val = o.create()
	o.opt = opt
}

// Set sets the *local* state of the wrapped value.
// It DOES NOT modify the store.
// This is implementing the pflag.Value interface.
func (o *OrgScoped[T]) Set(s string) error {
	return o.val.Set(s)
}

// Type returns the type of value this struct is describing.
// This is implementing the pflag.Value interface.
func (o *OrgScoped[T]) Type() string {
	return o.val.Type()
}

// String returns a string representation of the value this struct is describing.
// It is expected that this string representation can be used to unmarshal the value into the type T by calling
// Set() on this struct.
// This is implementing the pflag.Value interface.
func (o *OrgScoped[T]) String() string {
	return o.val.String()
}

// Identifier returns a stable string identifier which is used to store and retrieve org-level settings for this struct.
func (o *OrgScoped[T]) Identifier() string {
	return o.opt.Env
}

// Override upserts an org-level override for the given orgID, using this value's identifier as the key and the
// String() marshaling of the given value val.
func (o *OrgScoped[T]) Override(store Store, orgID uuid.UUID, val T) error {
	// TODO: val can be nil, which unsets the override; don't panic.
	return store.PutOrgSetting(orgID, o.Identifier(), val.String())
}

// GlobalValue returns the current value which has been set at the deployment-level.
func (o *OrgScoped[T]) GlobalValue() T {
	return o.val
}

// OrgValue returns an override - if set - or otherwise returns an error.
// If an override is present, it unmarshals the string representation from the store into a new instance of type T.
func (o *OrgScoped[T]) OrgValue(store Store, orgID uuid.UUID) (T, error) {
	var zero T

	val, err := store.GetOrgSetting(orgID, o.Identifier())
	if err != nil {
		return zero, err
	}

	// Override found, construct new instance of T.
	newInst := o.create()
	// val is a string, and Option.Set accepts a string to unmarshal into a complex type.
	if err = newInst.Set(val); err != nil {
		return zero, xerrors.Errorf("construct instance of %T: %w", newInst, err)
	}

	return newInst, nil
}

// Coalesce preferentially returns an org-level override if one has been set, otherwise it falls back to the
// deployment-level (global) value.
func (o *OrgScoped[T]) Coalesce(store Store, orgID uuid.UUID) (T, error) {
	val, err := o.OrgValue(store, orgID)
	if reflect.ValueOf(val).IsNil() {
		return o.GlobalValue(), nil
	}
	if err != nil {
		if errors.Is(err, EntryNotFound) {
			return o.GlobalValue(), nil
		}

		return val, err
	}

	return val, nil
}

// create instantiates a new instance of type T.
func (o *OrgScoped[T]) create() T {
	var zero T
	return reflect.New(reflect.TypeOf(zero).Elem()).Interface().(T)
}
