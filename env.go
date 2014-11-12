package env

// package main

import (
	"errors"
	"reflect"
)

// Env struct
type Env struct {
	value  reflect.Value // Value is the value of an interface or pointer
	prefix string
}

// Process takes a struct, and maps environment variables to its fields.
// Errors returned from underlying functions will bubble up to the surface.
func Process(v interface{}, prefix ...string) error {
	e, err := NewEnv(v, prefix...)
	if err != nil {
		return err
	}

	if err := e.Process(); err != nil {
		return err
	}

	return nil
}

// MustProcess maps environment variables to the fields of struct v.
// If any errors are returned, this function will panic.
func MustProcess(v interface{}, prefix ...string) {
	if err := Process(v, prefix...); err != nil {
		panic(err)
	}
}

var errInvalidValue = errors.New("expected value must be a pointer to a struct")

// NewEnv creates an Env and sets it value and optionally its prefix.
func NewEnv(v interface{}, prefix ...string) (*Env, error) {
	e := &Env{}

	if len(prefix) > 0 {
		e.SetPrefix(prefix[0])
	}

	if err := e.SetValue(v); err != nil {
		return e, err
	}

	e.SetValue(v)

	return e, nil
}

// SetValue sets Value of Env e
func (e *Env) SetValue(v interface{}) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errInvalidValue
	}
	if reflect.ValueOf(v).Elem().Kind() != reflect.Struct {
		return errInvalidValue
	}

	e.value = reflect.ValueOf(v).Elem()

	return nil
}

// SetPrefix sets prefix of Env e
func (e *Env) SetPrefix(prefix string) {
	e.prefix = prefix
}

// Process fills the Env's struct with values from environment
// variables and/or default values.
func (e *Env) Process() error {
	for _, name := range e.fieldNames() {
		field, _ := e.value.Type().FieldByName(name)
		v, err := getVal(e.prefix, field)

		if err != nil {
			return err
		}
		e.value.FieldByName(name).Set(v)
	}

	return nil
}

// fieldNames returns the name of all struct fields as a slice of strings
func (e *Env) fieldNames() []string {
	fieldType := e.value.Type()

	var fieldNames []string
	for i := 0; i < fieldType.NumField(); i++ {
		field := fieldType.Field(i)
		fieldNames = append(fieldNames, field.Name)
	}
	return fieldNames
}
