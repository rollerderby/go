package state

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/rollerderby/go/json"
)

type (
	ErrInvalidJSONType  error
	ErrInvalidJSONValue error
	ErrNoInitializer    error
	ErrNotImplemented   error
	ErrObjectKeys       error
	ErrExistingKey      error
	ErrInvalidEnum      error
	ErrNoKey            error
)

var (
	conversionTypes   = json.ValueType(255)
	errNoInitializer  = ErrNoInitializer(errors.New("No Initalizer"))
	errNotImplemented = ErrNotImplemented(errors.New("Not Implemented"))
	errNoKey          = ErrNotImplemented(errors.New("Key is missing"))
)

func errInvalidJSONType(j json.Value, expectedTypes ...json.ValueType) ErrInvalidJSONType {
	var eTypes []string
	var cTypes []string
	foundConversionTypes := false
	for _, t := range expectedTypes {
		if t == conversionTypes {
			foundConversionTypes = true
		} else if !foundConversionTypes {
			eTypes = append(eTypes, t.String())
		} else {
			cTypes = append(cTypes, t.String())
		}
	}
	if len(cTypes) == 0 {
		return ErrInvalidJSONType(fmt.Errorf("Invalid JSON Type: %v  Expected: %v", j.Type(), strings.Join(eTypes, ", ")))
	} else {
		return ErrInvalidJSONType(fmt.Errorf("Invalid JSON Type: %v  Expected: %v (Can convert from %v)", j.Type(), strings.Join(eTypes, ", "), strings.Join(cTypes, ", ")))
	}
}

func errInvalidJSONValue(j json.Value, err error) ErrInvalidJSONValue {
	return ErrInvalidJSONValue(fmt.Errorf("Invalid JSON Value: %v  Err: %v", j.JSON(false), err))
}

func errObjectKeys(j json.Value, missingKeys, extraKeys []string) ErrInvalidJSONValue {
	var buf bytes.Buffer
	buf.WriteString("Error setting object using JSON Value: ")
	buf.WriteString(j.JSON(false))

	if len(missingKeys) > 0 {
		buf.WriteString(", Missing Keys: [")
		buf.WriteString(strings.Join(missingKeys, ", "))
		buf.WriteString("]")
	}

	if len(extraKeys) > 0 {
		buf.WriteString(", Extra Keys: [")
		buf.WriteString(strings.Join(extraKeys, ", "))
		buf.WriteString("]")
	}

	return ErrObjectKeys(errors.New(buf.String()))
}

func errExistingKey(key string) ErrExistingKey {
	return ErrExistingKey(fmt.Errorf("Existing key in root: %q", key))
}

func errInvalidEnum(val string, values []string) ErrInvalidEnum {
	return ErrInvalidEnum(fmt.Errorf("Invalid enum %q, not in %q", val, values))
}
