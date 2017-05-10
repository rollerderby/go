package entity

//go:generate $GOPATH/bin/buildStates

import "github.com/rollerderby/go/logger"

var log = logger.New("entity")

func Initialize() error {
	log.Info("Initializing")

	if err := initializeState(); err != nil {
		log.Critf("Error initializing state: %v", err)
		return err
	}

	return nil
}
