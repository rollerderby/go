package state

import (
	"fmt"
	"strings"

	"github.com/rollerderby/go/json"
)

type Enum struct {
	values      []string
	value       string
	parent      Value
	path        string
	revision    uint64
	skipSave    bool
	saveNeeded  bool
	writeGroups []string
	readGroups  []string
}

func NewEnumOf(values ...string) func() Value {
	return func() Value {
		return &Enum{values: values}
	}
}

func (obj *Enum) Value() string {
	return obj.value
}

func (obj *Enum) SetValue(val string) error {
	if val == "" {
		if val != obj.value {
			obj.value = val
			Root.changedValue(obj)
		}
		return nil
	} else {
		val = strings.ToLower(val)
		for _, val2 := range obj.values {
			if strings.ToLower(val2) == val {
				if obj.value == val2 {
					return nil
				}
				obj.value = val2
				Root.changedValue(obj)
				return nil
			}
		}
		return errInvalidEnum(val, obj.values)
	}
}

func (obj *Enum) WriteGroups() []string {
	return obj.writeGroups
}
func (obj *Enum) AddWriteGroup(group ...string) {
	obj.writeGroups = mergeGroups(obj.writeGroups, group)
	obj.readGroups = mergeGroups(obj.readGroups, group)
}
func (obj *Enum) ReadGroups() []string {
	return obj.readGroups
}
func (obj *Enum) AddReadGroup(group ...string) {
	obj.readGroups = mergeGroups(obj.readGroups, group)
}
func (obj *Enum) SaveNeeded() bool       { return obj.saveNeeded }
func (obj *Enum) SetSaveNeeded(val bool) { obj.saveNeeded = val }
func (obj *Enum) SkipSave() bool         { return obj.skipSave }
func (obj *Enum) SetSkipSave(skip bool)  { obj.skipSave = skip }
func (obj *Enum) Revision() uint64       { return obj.revision }
func (obj *Enum) SetRevision(rev uint64) { obj.revision = rev }
func (obj *Enum) Path() string           { return obj.path }
func (obj *Enum) Parent() Value          { return obj.parent }
func (obj *Enum) String() string         { return fmt.Sprintf("%q", obj.value) }

func (obj *Enum) SetParentAndPath(parent Value, path string) {
	obj.parent = parent
	obj.path = path
	Root.changedValue(obj)
}

func (obj *Enum) JSON(skipSave bool) json.Value {
	return json.NewString(obj.value)
}

func (obj *Enum) SetJSON(j json.Value) error {
	if j == json.True {
		return obj.SetValue("true")
	} else if j == json.False {
		return obj.SetValue("false")
	} else if j == json.Null {
		return obj.SetValue("null")
	}

	switch j := j.(type) {
	case *json.String:
		return obj.SetValue(j.Get())
	case *json.Number:
		return obj.SetValue(j.String())
	default:
		return errInvalidJSONType(j, json.StringValue, conversionTypes, json.NumberValue, json.TrueValue, json.FalseValue, json.NullValue)
	}
}
