package state

import (
	"fmt"

	"github.com/rollerderby/go/json"
)

type Date struct {
	value       string
	parent      Value
	path        string
	revision    uint64
	skipSave    bool
	saveNeeded  bool
	writeGroups []string
	readGroups  []string
}

func NewDate() Value { return &Date{} }

func (obj *Date) Value() string {
	return obj.value
}

func (obj *Date) SetValue(val string) error {
	if obj.value == val {
		return nil
	}
	obj.value = val
	Root.changedValue(obj)
	return nil
}

func (obj *Date) WriteGroups() []string {
	return obj.writeGroups
}
func (obj *Date) AddWriteGroup(group ...string) {
	obj.writeGroups = mergeGroups(obj.writeGroups, group)
	obj.readGroups = mergeGroups(obj.readGroups, group)
}
func (obj *Date) ReadGroups() []string {
	return obj.readGroups
}
func (obj *Date) AddReadGroup(group ...string) {
	obj.readGroups = mergeGroups(obj.readGroups, group)
}
func (obj *Date) SaveNeeded() bool       { return obj.saveNeeded }
func (obj *Date) SetSaveNeeded(val bool) { obj.saveNeeded = val }
func (obj *Date) SkipSave() bool         { return obj.skipSave }
func (obj *Date) SetSkipSave(skip bool)  { obj.skipSave = skip }
func (obj *Date) Revision() uint64       { return obj.revision }
func (obj *Date) SetRevision(rev uint64) { obj.revision = rev }
func (obj *Date) Path() string           { return obj.path }
func (obj *Date) Parent() Value          { return obj.parent }
func (obj *Date) String() string         { return fmt.Sprintf("%q", obj.value) }

func (obj *Date) SetParentAndPath(parent Value, path string) {
	obj.parent = parent
	obj.path = path
	Root.changedValue(obj)
}

func (obj *Date) JSON(skipSave bool) json.Value {
	return json.NewString(obj.value)
}

func (obj *Date) SetJSON(j json.Value) error {
	switch j := j.(type) {
	case *json.String:
		return obj.SetValue(j.Get())
	default:
		return errInvalidJSONType(j, json.StringValue)
	}
}
