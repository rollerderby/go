package websocket

import "github.com/rollerderby/go/logger"

type User interface {
	Username() string
	Name() string
}

type Client interface {
	Close(error)
	User() User
	ExtraInfo() string
	Log() *logger.Logger
}
