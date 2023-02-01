// Copyright 2023 Urvantsev Evgenii. All rights reserved.
// Use of this source code is governed by a BSD3-style
// license that can be found in the LICENSE file.

package form

import (
	"errors"
	"reflect"
	"strconv"
)

var (
	errInvalidValue = errors.New("form: invalid value")
)

type InvalidLoadError struct {
	Type reflect.Type
}

func (e *InvalidLoadError) Error() string {
	if e.Type == nil {
		return "form: Parse(nil)"
	}

	if e.Type.Kind() != reflect.Pointer {
		return "form: Parse(non-pointer " + e.Type.String() + ")"
	}
	return "form: Parse(nil " + e.Type.String() + ")"
}

// An LoadTypeError describes a form value that was
// not appropriate for a value of a specific Go type.
type LoadTypeError struct {
	Value  string       // description of form value - "bool", "array", "number -5"
	Type   reflect.Type // type of Go value it could not be assigned to
	Struct string       // name of the struct type containing the field
	Field  string       // the full path from root node to the field
}

func (e *LoadTypeError) Error() string {
	if e.Struct != "" || e.Field != "" {
		return "form: cannot load " + e.Value + " into Go struct field " + e.Struct + "." + e.Field + " of type " + e.Type.String()
	}
	return "form: cannot load " + e.Value + " into Go value of type " + e.Type.String()
}

type decodeState struct {
	data       map[string][]string
	savedError error
}

func (d *decodeState) parse(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &InvalidLoadError{reflect.TypeOf(v)}
	}

	if err := d.value(rv); err != nil {
		return d.addErrorContext(err)
	}

	return d.savedError
}

func (d *decodeState) value(rv reflect.Value) error {
	v := rv.Elem()
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return errInvalidValue
	}

	fieldAliasNames := make([]string, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		aliasName := field.Tag.Get("request")
		if aliasName != "" {
			fieldAliasNames[i] = aliasName
			continue
		}

		fieldAliasNames[i] = field.Name
	}

	for i, fieldAliasName := range fieldAliasNames {
		dataV, ok := d.data[fieldAliasName]
		if !ok {
			continue
		}

		fieldValue := v.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		if fieldValue.Kind() == reflect.Slice {
			if fieldValue.Len() == 0 {
				v.Set(reflect.MakeSlice(v.Type(), len(dataV), len(dataV)))
			}

			for i := 0; i < fieldValue.Len(); i++ {
				fieldValueI := fieldValue.Index(i)
				switch fieldValue.Type().Elem().Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					intV, _ := strconv.ParseInt(dataV[i], 10, 64)
					fieldValueI.SetInt(intV)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
					intV, err := strconv.ParseUint(dataV[i], 10, 64)
					if err != nil {
						d.saveError(&LoadTypeError{Value: "array " + dataV[i], Type: v.Type()})
					}
					fieldValueI.SetUint(intV)
				case reflect.Float32, reflect.Float64:
					n, err := strconv.ParseFloat(dataV[i], fieldValueI.Type().Bits())
					if err != nil || fieldValueI.OverflowFloat(n) {
						d.saveError(&LoadTypeError{Value: "array " + dataV[i], Type: v.Type()})
						break
					}
					fieldValueI.SetFloat(n)
				case reflect.String, reflect.Interface:
					fieldValueI.SetString(dataV[i])
				}
			}
		}

		if len(dataV) < 1 {
			continue
		}

		if dataV[0] == "null" {
			continue
		}

		switch fieldValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intV, _ := strconv.ParseInt(dataV[0], 10, 64)
			fieldValue.SetInt(intV)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			intV, err := strconv.ParseUint(dataV[0], 10, 64)
			if err != nil {
				d.saveError(&LoadTypeError{Value: "number " + dataV[0], Type: v.Type()})
			}
			fieldValue.SetUint(intV)
		case reflect.Bool:
			v.SetBool(dataV[0] == "true" || dataV[0] == "1")
		case reflect.Float32, reflect.Float64:
			n, err := strconv.ParseFloat(dataV[0], fieldValue.Type().Bits())
			if err != nil || fieldValue.OverflowFloat(n) {
				d.saveError(&LoadTypeError{Value: "number " + dataV[0], Type: v.Type()})
				break
			}
			fieldValue.SetFloat(n)
		case reflect.String, reflect.Interface:
			fieldValue.SetString(dataV[0])
		default:
			d.savedError = errInvalidValue
		}
	}

	return nil
}

func (d *decodeState) saveError(err error) {
	if d.savedError == nil {
		d.savedError = d.addErrorContext(err)
	}
}

func (d *decodeState) init(data map[string][]string) {
	d.savedError = nil
	d.data = data
}

func (d *decodeState) addErrorContext(err error) error {
	return err
}
