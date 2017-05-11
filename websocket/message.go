package websocket

import (
	"errors"

	"github.com/rollerderby/go/json"
)

var ErrInvalidJSON = errors.New("Invalid JSON Format")

type JSON interface {
	JSON() json.Value
}

type Message struct {
	Type string
	Data json.Value
	ws   *Websocket
}

func (m *Message) JSON() json.Value {
	obj := make(json.Object)

	obj["type"] = json.NewString(m.Type)
	if m.Data != nil {
		obj["data"] = m.Data
	}

	return obj
}

func newMessage(jValue json.Value) (*Message, error) {
	obj, ok := jValue.(json.Object)
	if !ok {
		return nil, ErrInvalidJSON
	}

	msg := &Message{}
	if sObj, ok := obj["type"].(*json.String); !ok {
		return nil, ErrInvalidJSON
	} else {
		msg.Type = sObj.Get()
	}
	msg.Data = obj["data"]

	return msg, nil
}
