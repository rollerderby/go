package auth

import "github.com/rollerderby/go/state"

// Auto generated using buildStates command

var (
	Users *UsersHelper
)

func initializeState() error {
	state.Root.Lock()
	defer state.Root.Unlock()

	Users = newUsersHelper(newUsers().(*state.Hash))
	if err := state.Root.Add("User", "user", Users.state); err != nil {
		return err
	}
	return nil
}

func newUsers() state.Value {
	return state.NewHashOf(newUser)()
}

func newUser() state.Value {
	return &state.Object{
		Definition: state.ObjectDef{
			Name: "User",
			Values: []state.ObjectValueDef{
				state.ObjectValueDef{Name: "Username", Initializer: state.NewString},
				state.ObjectValueDef{Name: "PasswordHash", Initializer: state.NewString},
				state.ObjectValueDef{Name: "PasswordHashType", Initializer: state.NewEnumOf([]string{"sha512", "sha256", "md5"}...)},
				state.ObjectValueDef{Name: "IsSuper", Initializer: state.NewBool},
				state.ObjectValueDef{Name: "Groups", Initializer: state.NewArrayOf(state.NewString)},
				state.ObjectValueDef{Name: "PersonID", Initializer: state.NewGUID},
			}}}
}
