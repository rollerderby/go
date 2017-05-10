package state

import (
	"fmt"

	"github.com/rollerderby/go/json"
	uuid "github.com/satori/go.uuid"
)

type GUID struct {
	value      string
	parent     Value
	path       string
	revision   uint64
	skipSave   bool
	saveNeeded bool
}

func NewGUID() Value { return &GUID{} }

func (obj *GUID) SetNew() error {
	return obj.SetValue(uuid.NewV4().String())
}

func (obj *GUID) Value() string {
	return obj.value
}

func (obj *GUID) SetValue(val string) error {
	if val == "" {
		if val != obj.value {
			obj.value = val
			Root.changedValue(obj)
		}
	} else {
		guid, err := uuid.FromString(val)
		if err != nil {
			return err
		}
		if guid.String() == obj.value {
			return nil
		}
		obj.value = guid.String()
		Root.changedValue(obj)
	}
	return nil
}

func (obj *GUID) SaveNeeded() bool       { return obj.saveNeeded }
func (obj *GUID) SetSaveNeeded(val bool) { obj.saveNeeded = val }
func (obj *GUID) SkipSave() bool         { return obj.skipSave }
func (obj *GUID) SetSkipSave(skip bool)  { obj.skipSave = skip }
func (obj *GUID) Revision() uint64       { return obj.revision }
func (obj *GUID) SetRevision(rev uint64) { obj.revision = rev }
func (obj *GUID) Path() string           { return obj.path }
func (obj *GUID) Parent() Value          { return obj.parent }
func (obj *GUID) String() string         { return fmt.Sprintf("%q", obj.value) }

func (obj *GUID) SetParentAndPath(parent Value, path string) {
	obj.parent = parent
	obj.path = path
	Root.changedValue(obj)
}

func (obj *GUID) JSON(skipSave bool) json.Value {
	return json.NewString(obj.value)
}

func (obj *GUID) SetJSON(j json.Value) error {
	switch j := j.(type) {
	case *json.String:
		obj.SetValue(j.Get())
	default:
		return errInvalidJSONType(j, json.StringValue)
	}
	return nil
}
