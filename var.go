package env

import (
	"fmt"
	"os"
	"reflect"
	_ "regexp"
	"strconv"
	"strings"
)

type envVar struct {
	key       string
	keyPrefix string
	required  bool
	default_  reflect.Value
	options   []reflect.Value
}

// getVal returns a reflect.Value filled in with the associated
// environment variable or default.
func getVal(keyPrefix string, field reflect.StructField) (reflect.Value, error) {
	newVar := &envVar{keyPrefix: keyPrefix} //Default: reflect.ValueOf(nil)}
	newVar.parse(field)

	value, err := convert(field.Type, os.Getenv(newVar.key))
	if err != nil {
		return value, err
	}

	if value == reflect.ValueOf(nil) {
		if newVar.required {
			return value, fmt.Errorf("%s required", newVar.key)
		}

		// Check if we have a default value to set, otherwise set the type's zero value
		if newVar.default_ != reflect.ValueOf(nil) {
			value = newVar.default_
		} else {
			value = reflect.Zero(field.Type)
		}
	}

	if len(newVar.options) > 0 {
		if !newVar.optionsContains(value) {
			return value, fmt.Errorf(`%v="%v" not in allowed options: %v`, newVar.key, value, newVar.options)
		}
	}

	return value, nil
}

func (v *envVar) optionsContains(s reflect.Value) bool {
	for _, v := range v.options {
		if s.Interface() == v.Interface() {
			return true
		}
	}
	return false
}

// parse parses the struct tags of the given field.
func (v *envVar) parse(field reflect.StructField) error {
	key := field.Name

	tag := field.Tag.Get("env")

	if tag == "" {
		return nil
	}

	tagParams := strings.Split(tag, " ")
	for _, tagParam := range tagParams {
		var param, value string

		option := strings.Split(tagParam, "=")
		param = option[0]
		if len(option) > 1 {
			value = option[1]
		}

		switch param {
		case "key":
			key = value
		case "required":
			v.required = true
		case "default":
			d, err := convert(field.Type, value)
			if err != nil {
				return err
			}
			v.default_ = d
		case "options":
			in := strings.Split(value, ",")
			values := make([]reflect.Value, len(in))
			for k, val := range in {
				v1, err := convert(field.Type, val)
				if err != nil {
					return err
				}
				values[k] = v1
			}
			v.options = values
		}
	}

	v.key = strings.ToUpper(v.keyPrefix + key)

	return nil
}

// Convert a string into the specified type. Return the type's zero value
// if we receive an empty string
func convert(t reflect.Type, value string) (reflect.Value, error) {
	if value == "" {
		return reflect.ValueOf(nil), nil
	}

	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(value), nil
	case reflect.Int:
		return parseInt(value)
	case reflect.Bool:

		return parseBool(value)
	}

	return reflect.ValueOf(nil), conversionError(value, `unsupported `+t.Kind().String())
}

type errConversion struct {
	Value string
	Type  string
}

func (e *errConversion) Error() string {
	return fmt.Sprintf(`could not convert value "%s" into %s type`, e.Value, e.Type)
}

func conversionError(v, t string) *errConversion {
	return &errConversion{Value: v, Type: t}
}

func parseInt(value string) (reflect.Value, error) {
	if value == "" {
		return reflect.Zero(reflect.TypeOf(reflect.Int)), nil
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return reflect.ValueOf(nil), conversionError(value, "int")
	}
	return reflect.ValueOf(i), nil
}

func parseBool(value string) (reflect.Value, error) {
	if value == "" {
		return reflect.Zero(reflect.TypeOf(reflect.Int)), nil
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		return reflect.ValueOf(nil), conversionError(value, "bool")
	}
	return reflect.ValueOf(b), nil
}

func parseFloat(value string) (reflect.Value, error) {
	b, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return reflect.ValueOf(nil), conversionError(value, "float64")
	}
	return reflect.ValueOf(b), nil
}
