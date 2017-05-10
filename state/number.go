package state

import (
	"fmt"

	"github.com/rollerderby/go/json"
)

type Number struct {
	value      int64
	parent     Value
	path       string
	revision   uint64
	skipSave   bool
	saveNeeded bool
}

func NewNumber() Value { return &Number{} }

func (obj *Number) Value() int64 {
	return obj.value
}

func (obj *Number) SetValue(val int64) error {
	if obj.value == val {
		return nil
	}
	obj.value = val
	Root.changedValue(obj)
	return nil
}

func (obj *Number) SaveNeeded() bool       { return obj.saveNeeded }
func (obj *Number) SetSaveNeeded(val bool) { obj.saveNeeded = val }
func (obj *Number) SkipSave() bool         { return obj.skipSave }
func (obj *Number) SetSkipSave(skip bool)  { obj.skipSave = skip }
func (obj *Number) Revision() uint64       { return obj.revision }
func (obj *Number) SetRevision(rev uint64) { obj.revision = rev }
func (obj *Number) Path() string           { return obj.path }
func (obj *Number) Parent() Value          { return obj.parent }
func (obj *Number) String() string         { return fmt.Sprint(obj.value) }

func (obj *Number) SetParentAndPath(parent Value, path string) {
	obj.parent = parent
	obj.path = path
	Root.changedValue(obj)
}

func (obj *Number) JSON(skipSave bool) json.Value {
	return json.NewNumber(obj.value)
}

func (obj *Number) SetJSON(j json.Value) error {
	switch j := j.(type) {
	case *json.String:
		if val, err := j.AsNumber().GetInt64(); err != nil {
			return errInvalidJSONValue(j, err)
		} else {
			obj.SetValue(val)
		}
	case *json.Number:
		if val, err := j.GetInt64(); err != nil {
			return errInvalidJSONValue(j, err)
		} else {
			obj.SetValue(val)
		}
	default:
		return errInvalidJSONType(j, json.NumberValue, conversionTypes, json.StringValue)
	}
	return nil
}
