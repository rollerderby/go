package state

import (
	"fmt"
	"strings"

	"github.com/rollerderby/go/json"
)

type Array struct {
	initializer func() Value
	values      []Value
	parent      Value
	path        string
	revision    uint64
	skipSave    bool
	saveNeeded  bool
}

func NewArrayOf(initializer func() Value) func() Value {
	return func() Value {
		return &Array{initializer: initializer}
	}
}

func (obj *Array) NewEmptyElement() (Value, error) {
	return obj.NewElement(nil)
}

func (obj *Array) NewElement(j json.Value) (Value, error) {
	if obj.initializer == nil {
		return nil, errNoInitializer
	}

	elem := obj.initializer()
	if j != nil {
		if err := elem.SetJSON(j); err != nil {
			return nil, err
		}
	}

	if obj.parent != nil {
		elem.SetParentAndPath(obj, fmt.Sprintf("%v[%v]", obj.path, len(obj.values)))
	}
	obj.values = append(obj.values, elem)
	Root.changedValue(obj)
	return elem, nil
}

func (obj *Array) Clear() {
	for _, value := range obj.values {
		value.SetParentAndPath(nil, "")
	}
	obj.values = nil
	Root.changedValue(obj)
}

func (obj *Array) Values() []Value {
	var ret []Value
	ret = append(ret, obj.values...)
	return ret
}

func (obj *Array) SaveNeeded() bool       { return obj.saveNeeded }
func (obj *Array) SetSaveNeeded(val bool) { obj.saveNeeded = val }
func (obj *Array) SkipSave() bool         { return obj.skipSave }
func (obj *Array) SetSkipSave(skip bool)  { obj.skipSave = skip }
func (obj *Array) Revision() uint64       { return obj.revision }
func (obj *Array) SetRevision(rev uint64) { obj.revision = rev }
func (obj *Array) Path() string           { return obj.path }
func (obj *Array) Parent() Value          { return obj.parent }
func (obj *Array) String() string {
	var children []string
	for _, value := range obj.values {
		children = append(children, value.String())
	}
	return fmt.Sprintf("[%v]", strings.Join(children, ", "))
}

func (obj *Array) SetParentAndPath(parent Value, path string) {
	obj.parent = parent
	obj.path = path

	for idx, value := range obj.values {
		if parent != nil {
			value.SetParentAndPath(obj, fmt.Sprintf("%v[%v]", obj.path, idx))
		} else {
			value.SetParentAndPath(nil, "")
		}
	}

	Root.changedValue(obj)
}

func (obj *Array) JSON(skipSave bool) json.Value {
	var j json.Array
	for _, value := range obj.values {
		if !skipSave || !value.SkipSave() {
			j = append(j, value.JSON(skipSave))
		}
	}
	return j
}

func (obj *Array) SetJSON(j json.Value) error {
	jArray, ok := j.(json.Array)
	if !ok {
		return errInvalidJSONType(j, json.ArrayValue)
	}

	obj.Clear()
	for _, jValue := range jArray {
		if _, err := obj.NewElement(jValue); err != nil {
			return err
		}
	}
	Root.changedValue(obj)
	return nil
}
