package state

import "github.com/rollerderby/go/json"

type Value interface {
	Revision() uint64
	SetRevision(rev uint64)

	Parent() Value
	Path() string
	SetParentAndPath(parent Value, path string)

	JSON(skipSave bool) json.Value
	SetJSON(j json.Value) error

	SkipSave() bool
	SetSkipSave(skip bool)

	SaveNeeded() bool
	SetSaveNeeded(val bool)

	String() string
}
