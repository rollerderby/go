package state

import (
	"fmt"
	"strings"

	"github.com/rollerderby/go/json"
)

type ObjectDef struct {
	Name   string
	Values []ObjectValueDef
}

type ObjectValueDef struct {
	Name        string
	Initializer func() Value
}

type Object struct {
	Definition      ObjectDef
	AllowPartialSet bool
	IgnoreExtraData bool
	values          map[string]Value
	parent          Value
	path            string
	revision        uint64
	skipSave        bool
	saveNeeded      bool
}

func (obj *Object) init() {
	if obj.values != nil {
		return
	}

	obj.values = make(map[string]Value)
	for _, value := range obj.Definition.Values {
		obj.values[value.Name] = value.Initializer()
	}
}

func (obj *Object) Get(key string) Value {
	obj.init()
	return obj.values[key]
}

func (obj *Object) SaveNeeded() bool       { return obj.saveNeeded }
func (obj *Object) SetSaveNeeded(val bool) { obj.saveNeeded = val }
func (obj *Object) SkipSave() bool         { return obj.skipSave }
func (obj *Object) SetSkipSave(skip bool)  { obj.skipSave = skip }
func (obj *Object) Revision() uint64       { return obj.revision }
func (obj *Object) SetRevision(rev uint64) { obj.revision = rev }
func (obj *Object) Path() string           { return obj.path }
func (obj *Object) Parent() Value          { return obj.parent }
func (obj *Object) String() string {
	var children []string
	obj.init()
	for _, value := range obj.Definition.Values {
		children = append(children, fmt.Sprintf("%v: %v", value.Name, obj.Get(value.Name).String()))
	}
	return fmt.Sprintf("{%v}", strings.Join(children, ", "))
}

func (obj *Object) SetParentAndPath(parent Value, path string) {
	obj.parent = parent
	obj.path = path

	obj.init()
	for idx, value := range obj.values {
		if parent != nil {
			value.SetParentAndPath(obj, fmt.Sprintf("%v[%v]", obj.path, idx))
		} else {
			value.SetParentAndPath(nil, "")
		}
	}
	Root.changedValue(obj)
}

func (obj *Object) JSON(skipSave bool) json.Value {
	j := make(json.Object)
	obj.init()
	for _, value := range obj.Definition.Values {
		val := obj.Get(value.Name)
		if !skipSave || !val.SkipSave() {
			j[value.Name] = val.JSON(skipSave)
		}
	}
	return j
}

func (obj *Object) SetJSON(j json.Value) error {
	object, ok := j.(json.Object)
	if !ok {
		return errInvalidJSONType(j, json.ObjectValue)
	}

	obj.init()
	var missingKeys, extraKeys []string
	if !obj.AllowPartialSet {
		for _, value := range obj.Definition.Values {
			if _, ok := object[value.Name]; !ok {
				missingKeys = append(missingKeys, value.Name)
			}
		}
	}
	if !obj.IgnoreExtraData {
		for key, _ := range object {
			if _, ok := obj.values[key]; !ok {
				extraKeys = append(extraKeys, key)
			}
		}
	}
	if len(missingKeys) > 0 || len(extraKeys) > 0 {
		return errObjectKeys(j, missingKeys, extraKeys)
	}
	for _, value := range obj.Definition.Values {
		if jValue, ok := object[value.Name]; ok {
			if err := obj.values[value.Name].SetJSON(jValue); err != nil {
				return err
			}
		}
	}
	Root.changedValue(obj)
	return nil
}
