package auth

import "github.com/rollerderby/go/state"

// Auto generated using buildStates command

type UsersHelper struct {
	state *state.Hash
}

func newUsersHelper(state *state.Hash) *UsersHelper {
	ret := &UsersHelper{state}
	return ret
}

func (h *UsersHelper) Path() string {
	return h.state.Path()
}

func (h *UsersHelper) JSON(indent bool) string {
	return h.state.JSON(false).JSON(indent)
}

func (h *UsersHelper) New(key string) (*UserHelper, error) {
	stateObj, err := h.state.NewEmptyElement(key)
	if err != nil {
		return nil, err
	}
	return newUserHelper(stateObj.(*state.Object)), nil
}

func (h *UsersHelper) Values() []*UserHelper {
	var ret []*UserHelper
	for _, val := range h.state.Values() {
		ret = append(ret, newUserHelper(val.(*state.Object)))
	}
	return ret
}

func (h *UsersHelper) Keys() []string {
	return h.state.Keys()
}

func (h *UsersHelper) FindByUsername(lookFor string) []*UserHelper {
	var ret []*UserHelper
	for _, obj := range h.Values() {
		if obj.Username() == lookFor {
			ret = append(ret, obj)
		}
	}
	return ret
}

func (h *UsersHelper) FindByPasswordHash(lookFor string) []*UserHelper {
	var ret []*UserHelper
	for _, obj := range h.Values() {
		if obj.PasswordHash() == lookFor {
			ret = append(ret, obj)
		}
	}
	return ret
}

func (h *UsersHelper) FindByPasswordHashType(lookFor string) []*UserHelper {
	var ret []*UserHelper
	for _, obj := range h.Values() {
		if obj.PasswordHashType() == lookFor {
			ret = append(ret, obj)
		}
	}
	return ret
}

func (h *UsersHelper) FindByIsSuper(lookFor bool) []*UserHelper {
	var ret []*UserHelper
	for _, obj := range h.Values() {
		if obj.IsSuper() == lookFor {
			ret = append(ret, obj)
		}
	}
	return ret
}

func (h *UsersHelper) FindByPersonID(lookFor string) []*UserHelper {
	var ret []*UserHelper
	for _, obj := range h.Values() {
		if obj.PersonID() == lookFor {
			ret = append(ret, obj)
		}
	}
	return ret
}

type UserHelper struct {
	state *state.Object
}

func newUserHelper(state *state.Object) *UserHelper {
	ret := &UserHelper{state}
	return ret
}

func (h *UserHelper) Path() string {
	return h.state.Path()
}

func (h *UserHelper) JSON(indent bool) string {
	return h.state.JSON(false).JSON(indent)
}

func (h *UserHelper) Username() string {
	return h.state.Get("Username").(*state.String).Value()
}

func (h *UserHelper) SetUsername(val string) error {
	return h.state.Get("Username").(*state.String).SetValue(val)
}

func (h *UserHelper) PasswordHash() string {
	return h.state.Get("PasswordHash").(*state.String).Value()
}

func (h *UserHelper) SetPasswordHash(val string) error {
	return h.state.Get("PasswordHash").(*state.String).SetValue(val)
}

func (h *UserHelper) PasswordHashType() string {
	return h.state.Get("PasswordHashType").(*state.Enum).Value()
}

func (h *UserHelper) SetPasswordHashType(val string) error {
	return h.state.Get("PasswordHashType").(*state.Enum).SetValue(val)
}

func (h *UserHelper) IsSuper() bool {
	return h.state.Get("IsSuper").(*state.Bool).Value()
}

func (h *UserHelper) SetIsSuper(val bool) error {
	return h.state.Get("IsSuper").(*state.Bool).SetValue(val)
}

func (h *UserHelper) Groups() *User_GroupsHelper {
	return newUser_GroupsHelper(h.state.Get("Groups").(*state.Array))
}
func (h *UserHelper) PersonID() string {
	return h.state.Get("PersonID").(*state.GUID).Value()
}

func (h *UserHelper) SetPersonID(val string) error {
	return h.state.Get("PersonID").(*state.GUID).SetValue(val)
}

type User_GroupsHelper struct {
	state *state.Array
}

func newUser_GroupsHelper(state *state.Array) *User_GroupsHelper {
	ret := &User_GroupsHelper{state}
	return ret
}

func (h *User_GroupsHelper) Path() string {
	return h.state.Path()
}

func (h *User_GroupsHelper) JSON(indent bool) string {
	return h.state.JSON(false).JSON(indent)
}

func (h *User_GroupsHelper) Clear() {
	h.state.Clear()
}

func (h *User_GroupsHelper) Add(v string) error {
	elem, err := h.state.NewEmptyElement()
	if err != nil {
		return err
	}
	return elem.(*state.String).SetValue(v)
}

func (h *User_GroupsHelper) New() (*state.String, error) {
	elem, err := h.state.NewEmptyElement()
	if err != nil {
		return nil, err
	}
	return elem.(*state.String), nil
}

func (h *User_GroupsHelper) Values() []string {
	var ret []string
	for _, val := range h.state.Values() {
		ret = append(ret, val.(*state.String).Value())
	}
	return ret
}
