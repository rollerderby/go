package server

import (
	"net/http"

	"github.com/rollerderby/go/auth"
	"github.com/rollerderby/go/logger"
	"github.com/rollerderby/go/websocket"
)

type controlConnection struct {
	ws   *websocket.Websocket
	user *auth.User
}

func controlHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	c := &controlConnection{user: auth.CheckAuth(r)}

	if c.ws, err = websocket.New(c, w, r); err != nil {
		log.Errf("Cannot make websocket: %v", err)
		return
	}

	c.menuItems(nil)

	c.ws.Register("MenuItems", c.menuItems)
	c.ws.Loop()
}

func (c *controlConnection) Close(err error) {
}

func (c *controlConnection) User() websocket.User {
	return c.user
}

func (c *controlConnection) ExtraInfo() string {
	return ""
}

func (c *controlConnection) Log() *logger.Logger {
	return log
}

func (c *controlConnection) menuItems(msg *websocket.Message) error {
	c.ws.SendResponse("MenuItems", auth.MenuItems(c.user))
	return nil
}
