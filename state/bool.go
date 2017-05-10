package state

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rollerderby/go/json"
)

type Bool struct {
	value      bool
	parent     Value
	path       string
	revision   uint64
	skipSave   bool
	saveNeeded bool
}

func NewBool() Value { return &Bool{} }

func (obj *Bool) Value() bool {
	return obj.value
}

func (obj *Bool) SetValue(val bool) error {
	if obj.value == val {
		return nil
	}
	obj.value = val
	Root.changedValue(obj)
	return nil
}

func (obj *Bool) SaveNeeded() bool       { return obj.saveNeeded }
func (obj *Bool) SetSaveNeeded(val bool) { obj.saveNeeded = val }
func (obj *Bool) SkipSave() bool         { return obj.skipSave }
func (obj *Bool) SetSkipSave(skip bool)  { obj.skipSave = skip }
func (obj *Bool) Revision() uint64       { return obj.revision }
func (obj *Bool) SetRevision(rev uint64) { obj.revision = rev }
func (obj *Bool) Path() string           { return obj.path }
func (obj *Bool) Parent() Value          { return obj.parent }
func (obj *Bool) String() string         { return fmt.Sprint(obj.value) }

func (obj *Bool) SetParentAndPath(parent Value, path string) {
	obj.parent = parent
	obj.path = path
	Root.changedValue(obj)
}

func (obj *Bool) JSON(skipSave bool) json.Value {
	if obj.value {
		return json.True
	}
	return json.False
}

func (obj *Bool) SetJSON(j json.Value) error {
	if j == json.True {
		obj.SetValue(true)
		return nil
	} else if j == json.False {
		obj.SetValue(false)
		return nil
	}

	switch j := j.(type) {
	case *json.String:
		// Try to figure out value
		val := strings.TrimSpace(strings.ToLower(j.Get()))
		if val == "true" {
			obj.SetValue(true)
			return nil
		} else if val == "false" {
			obj.SetValue(false)
			return nil
		}

		num := j.AsNumber()
		if val, err := num.GetInt64(); err == nil {
			obj.SetValue(val != 0)
			return nil
		}
		return errInvalidJSONValue(j, errors.New("Not a bool string"))
	case *json.Number:
		// zero = false
		// else true
		if val, err := j.GetInt64(); err != nil {
			return errInvalidJSONValue(j, err)
		} else {
			obj.SetValue(val != 0)
		}
	default:
		return errInvalidJSONType(j, json.TrueValue, json.FalseValue, conversionTypes, json.StringValue, json.NumberValue)
	}
	return nil
}
