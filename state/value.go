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

	WriteGroups() []string
	AddWriteGroup(group ...string)
	ReadGroups() []string
	AddReadGroup(group ...string)
}

func mergeGroups(a, b []string) []string {
	var ret []string

	merge := func(in []string) {
		for _, c := range in {
			found := false
			for _, d := range ret {
				if d == c {
					found = true
					break
				}
			}
			if !found {
				ret = append(ret, c)
			}
		}
	}
	merge(a)
	merge(b)

	return ret
}
