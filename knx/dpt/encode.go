package dpt

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func EncodeDPTFromStringN(dptName, value string) ([]byte, error) {
	dv, ok := Produce(dptName)
	if !ok {
		return nil, fmt.Errorf("DPT not supported: %s", dptName)
	}
	return EncodeDPTFromString(dv, value)
}

func EncodeDPTFromString(dv DatapointValue, value string) ([]byte, error) {
	rv := reflect.ValueOf(dv)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return nil, fmt.Errorf("unexpected datapoint value: %s, %T", rv.Kind(), dv)
	}

	elem := rv.Elem()
	if !elem.CanSet() {
		return nil, fmt.Errorf("the value %s can not be set for %T", value, dv)
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return nil, errors.New("value must be provided")
	}

	switch elem.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("bool not valid %q: %w", value, err)
		}
		elem.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, elem.Type().Bits())
		if err != nil {
			return nil, fmt.Errorf("int not valid %q: %w", value, err)
		}
		if elem.OverflowInt(i) {
			return nil, fmt.Errorf("overflow int %q for %s", value, elem.Type())
		}
		elem.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u, err := strconv.ParseUint(value, 10, elem.Type().Bits())
		if err != nil {
			return nil, fmt.Errorf("uint not valid %q: %w", value, err)
		}
		if elem.OverflowUint(u) {
			return nil, fmt.Errorf("overflow uint %q for %s", value, elem.Type())
		}
		elem.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, elem.Type().Bits())
		if err != nil {
			return nil, fmt.Errorf("float not valid %q: %w", value, err)
		}
		elem.SetFloat(f)

	case reflect.String:
		elem.SetString(value)

	default:
		return nil, fmt.Errorf("kind %s not managed for kind %T", elem.Kind(), dv)
	}

	return dv.Pack(), nil
}
