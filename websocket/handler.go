package websocket

import "strings"

type handler struct {
	t string
	f func(*Message) error
}

func (ws *Websocket) Register(t string, f func(*Message) error) {
	t = strings.ToLower(t)
	h := &handler{t, f}
	ws.handlers = append(ws.handlers, h)
}
