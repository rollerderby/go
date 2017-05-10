package state

import (
	"fmt"

	"github.com/rollerderby/go/json"
)

type String struct {
	value      string
	parent     Value
	path       string
	revision   uint64
	skipSave   bool
	saveNeeded bool
}

func NewString() Value { return &String{} }

func (obj *String) Value() string {
	return obj.value
}

func (obj *String) SetValue(val string) error {
	if obj.value == val {
		return nil
	}
	obj.value = val
	Root.changedValue(obj)
	return nil
}

func (obj *String) SaveNeeded() bool       { return obj.saveNeeded }
func (obj *String) SetSaveNeeded(val bool) { obj.saveNeeded = val }
func (obj *String) SkipSave() bool         { return obj.skipSave }
func (obj *String) SetSkipSave(skip bool)  { obj.skipSave = skip }
func (obj *String) Revision() uint64       { return obj.revision }
func (obj *String) SetRevision(rev uint64) { obj.revision = rev }
func (obj *String) Path() string           { return obj.path }
func (obj *String) Parent() Value          { return obj.parent }
func (obj *String) String() string         { return fmt.Sprintf("%q", obj.value) }

func (obj *String) SetParentAndPath(parent Value, path string) {
	obj.parent = parent
	obj.path = path
	Root.changedValue(obj)
}

func (obj *String) JSON(skipSave bool) json.Value {
	return json.NewString(obj.value)
}

func (obj *String) SetJSON(j json.Value) error {
	if j == json.True {
		obj.SetValue("true")
		return nil
	} else if j == json.False {
		obj.SetValue("false")
		return nil
	} else if j == json.Null {
		obj.SetValue("null")
		return nil
	}

	switch j := j.(type) {
	case *json.String:
		obj.SetValue(j.Get())
	case *json.Number:
		obj.SetValue(j.String())
	default:
		return errInvalidJSONType(j, json.StringValue, conversionTypes, json.NumberValue, json.TrueValue, json.FalseValue, json.NullValue)
	}
	return nil
}
