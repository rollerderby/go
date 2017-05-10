package websocket

import "github.com/rollerderby/go/logger"

var log *logger.Logger = logger.New("websocket")

func Initialize() {
	log.Info("Initializing")
}
