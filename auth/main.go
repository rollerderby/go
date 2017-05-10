package auth

import (
	"github.com/rollerderby/go/entity"
	"github.com/rollerderby/go/logger"
	"github.com/rollerderby/go/state"
)

//go:generate $GOPATH/bin/buildStates

var log = logger.New("Auth")

func Initialize() error {
	log.Info("Initializing")

	initializeState()

	return nil
}

func AddGlobalUsers() error {
	state.Root.Lock()
	defer state.Root.Unlock()

	defaultUsers := []struct {
		username string
		password string
		isSuper  bool
		groups   []string
	}{
		{"admin", "admin", true, nil},
		{"readonly", "readonly", false, []string{"readonly"}},
	}

	var err error

	for _, u := range defaultUsers {
		var person *entity.PersonHelper

		for _, per := range entity.People.FindByName(u.username) {
			person = per
			break
		}

		if person == nil {
			person, err = entity.People.New("")
			if err != nil {
				log.Errorf("Cannot create person: %v", err)
				return err
			}
			person.SetName(u.username)
		}

		users := Users.FindByPersonID(person.ID())
		if len(users) == 0 {
			_, err := Users.AddUser(u.username, u.password, u.isSuper, u.groups, person.ID())
			return err
		}
	}
	return nil
}
