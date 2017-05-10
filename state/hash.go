package state

import (
	"fmt"
	"strings"

	"github.com/rollerderby/go/json"
)

type Hash struct {
	initializer func() Value
	values      map[string]Value
	isIDObject  bool
	parent      Value
	path        string
	revision    uint64
	skipSave    bool
	saveNeeded  bool
}

func NewHashOf(initializer func() Value) func() Value {
	isIDObject := false
	if testObj, ok := initializer().(*Object); ok {
		if len(testObj.Definition.Values) > 0 {
			valDef := testObj.Definition.Values[0]
			if valDef.Name == "ID" {
				val := valDef.Initializer()
				if _, ok := val.(*GUID); ok {
					isIDObject = true
				}
			}
		}
	}

	return func() Value {
		return &Hash{initializer: initializer, isIDObject: isIDObject}
	}
}

func (obj *Hash) init() {
	if obj.values != nil {
		return
	}

	obj.values = make(map[string]Value)
}

func (obj *Hash) NewEmptyElement(key string) (Value, error) {
	return obj.NewElement(key, nil)
}

func (obj *Hash) NewElement(key string, jValue json.Value) (Value, error) {
	if obj.initializer == nil {
		return nil, errNoInitializer
	}
	if key == "" && !obj.isIDObject {
		return nil, errNoKey
	}

	elem := obj.initializer()
	if jValue != nil {
		if err := elem.SetJSON(jValue); err != nil {
			return nil, err
		}
	}

	if obj.isIDObject {
		elemObj := elem.(*Object)
		idValue := elemObj.Get("ID").(*GUID)
		if key != "" {
			// Set ID on object to ensure they are the same
			// unless the set fails, in which case set the key
			// to the ID on object
			if err := idValue.SetValue(key); err != nil {
				key = idValue.Value()
			}
		} else {
			// Set key to ID on object
			key = idValue.Value()
		}

		if key == "" {
			idValue.SetNew()
			key = idValue.Value()
		}
	}

	if obj.parent != nil {
		elem.SetParentAndPath(obj, fmt.Sprintf("%v[%v]", obj.path, key))
	}
	obj.init()
	if oldValue, ok := obj.values[key]; ok {
		oldValue.SetParentAndPath(nil, "")
	}
	obj.values[key] = elem
	Root.changedValue(obj)
	return elem, nil
}

func (obj *Hash) Clear() {
	for _, value := range obj.values {
		value.SetParentAndPath(nil, "")
	}
	obj.values = nil
	Root.changedValue(obj)
}

func (obj *Hash) Keys() []string {
	var ret []string
	obj.init()
	for key, _ := range obj.values {
		ret = append(ret, key)
	}
	return ret
}

func (obj *Hash) Values() []Value {
	var ret []Value
	obj.init()
	for _, value := range obj.values {
		ret = append(ret, value)
	}
	return ret
}

func (obj *Hash) Get(key string) Value {
	obj.init()
	return obj.values[key]
}

func (obj *Hash) SaveNeeded() bool       { return obj.saveNeeded }
func (obj *Hash) SetSaveNeeded(val bool) { obj.saveNeeded = val }
func (obj *Hash) SkipSave() bool         { return obj.skipSave }
func (obj *Hash) SetSkipSave(skip bool)  { obj.skipSave = skip }
func (obj *Hash) Revision() uint64       { return obj.revision }
func (obj *Hash) SetRevision(rev uint64) { obj.revision = rev }
func (obj *Hash) Path() string           { return obj.path }
func (obj *Hash) Parent() Value          { return obj.parent }
func (obj *Hash) String() string {
	var children []string
	obj.init()
	for key, value := range obj.values {
		children = append(children, fmt.Sprintf("%v: %v", key, value.String()))
	}
	return fmt.Sprintf("{%v}", strings.Join(children, ", "))
}

func (obj *Hash) SetParentAndPath(parent Value, path string) {
	obj.parent = parent
	obj.path = path

	obj.init()
	for key, value := range obj.values {
		if parent != nil {
			value.SetParentAndPath(obj, fmt.Sprintf("%v[%v]", obj.path, key))
		} else {
			value.SetParentAndPath(nil, "")
		}
	}
	Root.changedValue(obj)
}

func (obj *Hash) JSON(skipSave bool) json.Value {
	j := make(json.Object)
	obj.init()
	for key, value := range obj.values {
		if !skipSave || !value.SkipSave() {
			j[key] = value.JSON(skipSave)
		}
	}
	return j
}

func (obj *Hash) SetJSON(j json.Value) error {
	jObject, ok := j.(json.Object)
	if !ok {
		return errInvalidJSONType(j, json.ObjectValue)
	}

	obj.init()
	for key, jValue := range jObject {
		if val := obj.Get(key); val != nil {
			if err := val.SetJSON(jValue); err != nil {
				return err
			}
		} else {
			if _, err := obj.NewElement(key, jValue); err != nil {
				return err
			}
		}
	}
	return nil
}
