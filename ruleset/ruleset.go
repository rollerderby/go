package ruleset

//go:generate $GOPATH/bin/buildStates

import (
	"github.com/rollerderby/go/logger"
	"github.com/rollerderby/go/state"
)

var log = logger.New("ruleset")

func Initialize() error {
	log.Info("Initializing")

	initializeState()

	state.Root.Lock()
	defer state.Root.Unlock()
	std, err := Rulesets.New("00000000-0000-0000-0000-000000000000")
	if err != nil {
		return err
	}
	std.SetLocked(true)

	return nil
}
