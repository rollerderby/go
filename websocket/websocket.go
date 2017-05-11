package websocket

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	gws "github.com/gorilla/websocket"
	"github.com/rollerderby/go/json"
	"github.com/rollerderby/go/logger"
)

var websockets []*Websocket
var mux sync.Mutex

type PacketInfo struct {
	packets int64
	bytes   int64
}

func (pi *PacketInfo) String() string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	unitIdx := 0
	size := float64(pi.bytes)
	for size > 1024 {
		size /= 1024.0
		unitIdx++
	}
	if unitIdx == 0 {
		return fmt.Sprintf("%v Packets, %vB", pi.packets, pi.bytes)
	}
	return fmt.Sprintf("%v Packets, %0.1f%v", pi.packets, size, units[unitIdx])
}

type Websocket struct {
	sync.Mutex // Protects against concurrent writes (if option enabled)

	client     Client
	log        *logger.Logger
	conn       *gws.Conn
	handlers   []*handler
	path       string
	sent       PacketInfo
	recv       PacketInfo
	lastActive time.Time
	LockWrites bool
}

type WebsocketInfo struct {
	Path       string
	Username   string
	Fullname   string
	Sent       string
	Recv       string
	ExtraInfo  string
	RemoteAddr string
	LastActive string
}

var checkOriginUpgrader = &gws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var noCheckOriginUpgrader = &gws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func register(ws *Websocket) {
	mux.Lock()
	defer mux.Unlock()

	log.Debugf("Register: %v %v", ws.path, ws.conn.RemoteAddr())
	websockets = append(websockets, ws)
}

func unregister(ws *Websocket) {
	mux.Lock()
	defer mux.Unlock()

	log.Debugf("Unregister: %v %v", ws.path, ws.conn.RemoteAddr())
	for i, ws2 := range websockets {
		if ws == ws2 {
			if len(websockets) <= 1 {
				websockets = nil
			} else {
				websockets[i] = websockets[len(websockets)-1]
				websockets = websockets[:len(websockets)-1]
			}
			break
		}
	}
}

func WebsocketInfos() []*WebsocketInfo {
	mux.Lock()
	defer mux.Unlock()

	var ret []*WebsocketInfo
	for _, ws := range websockets {
		username, fullname := "", ""
		if user := ws.client.User(); user != nil {
			username, fullname = user.Username(), user.Name()
		}
		ret = append(ret, &WebsocketInfo{
			Path:       ws.path,
			Username:   username,
			Fullname:   fullname,
			Sent:       ws.sent.String(),
			Recv:       ws.recv.String(),
			ExtraInfo:  ws.client.ExtraInfo(),
			RemoteAddr: ws.conn.RemoteAddr().String(),
			LastActive: ws.lastActive.Format(time.RFC3339),
		})
	}
	return ret
}

func New(client Client, w http.ResponseWriter, r *http.Request) (*Websocket, error) {
	return _new(checkOriginUpgrader, client, w, r)
}

func NewWithoutCheckOrgin(client Client, w http.ResponseWriter, r *http.Request) (*Websocket, error) {
	return _new(noCheckOriginUpgrader, client, w, r)
}

func _new(upgrader *gws.Upgrader, client Client, w http.ResponseWriter, r *http.Request) (*Websocket, error) {
	var err error
	if client == nil {
		return nil, errors.New("Client cannot be nil")
	}

	parentLog := client.Log()
	if parentLog == nil {
		parentLog = log
	}

	ws := &Websocket{client: client, log: parentLog.Child("WS")}
	if ws.conn, err = upgrader.Upgrade(w, r, nil); err != nil {
		return nil, err
	}

	ws.path = r.URL.Path
	register(ws)

	return ws, nil
}

func (ws *Websocket) Debug(d bool) {
	if d {
		ws.log.SetLevel(logger.DEBUG)
	} else {
		ws.log.SetLevel(logger.INFO)
	}
}

func (ws *Websocket) RemoteAddr() string {
	if ws != nil || ws.conn != nil {
		return ws.conn.RemoteAddr().String()
	}
	return "DISCONNECTED"
}

func (ws *Websocket) SendError(msg string, err error) error {
	ws.log.Debugf("%v  Error:  msg %q  err %v", ws.conn.RemoteAddr(), msg, err)
	return ws.sendMessage(&Message{Type: "Error", Data: json.NewString(msg)})
}

func (ws *Websocket) SendResponse(t string, data JSON) error {
	if data == nil {
		return ws.sendMessage(&Message{Type: t})
	}
	return ws.sendMessage(&Message{Type: t, Data: data.JSON()})
}

func (ws *Websocket) writeJSON(v json.Value) error {
	data := v.JSON(false)

	if ws.LockWrites {
		ws.Lock()
		defer ws.Unlock()
	}

	if err := ws.conn.WriteMessage(gws.TextMessage, []byte(data)); err != nil {
		return err
	}

	ws.sent.packets++
	ws.sent.bytes += int64(len(data))
	return nil
}

func (ws *Websocket) sendMessage(msg *Message) error {
	if strings.ToLower(msg.Type) != "pong" {
		ws.log.Debugf("%v  Sending %+v", ws.conn.RemoteAddr(), msg)
	}

	err := ws.writeJSON(msg.JSON())
	if err != nil {
		ws.log.Errorf("%v  Could not write message: %v", ws.conn.RemoteAddr(), err)
		return err
	}

	ws.lastActive = time.Now()
	return nil
}

func (ws *Websocket) Close() {
	if ws.client != nil {
		ws.client.Close(nil)
		ws.client = nil
	}

	if ws.conn != nil {
		unregister(ws)
		ws.conn.Close()
		ws.conn = nil
	}
}

func (ws *Websocket) Loop() {
	defer ws.Close()
	for {
		ws.conn.SetReadDeadline(time.Now().Add(time.Duration(10) * time.Minute))
		messageType, p, err := ws.conn.ReadMessage()
		if err == nil && messageType != gws.TextMessage {
			err = errors.New("Expected Text Message")
		}
		if err != nil {
			if _, ok := err.(*gws.CloseError); ok {
				err = nil
			} else if err, ok := err.(net.Error); ok && err.Timeout() {
				err = nil
			} else {
				ws.log.Errorf("%v  Error: %v", ws.conn.RemoteAddr(), err)
			}
			ws.client.Close(err)
			ws.client = nil
			return
		}
		ws.recv.packets++
		ws.recv.bytes += int64(len(p))

		jValue, err := json.Decode(p)
		if err != nil {
			ws.log.Errorf("%v  Error: %v", ws.conn.RemoteAddr(), err)
			ws.client.Close(err)
			ws.client = nil
			return
		}

		msg, err := newMessage(jValue)
		if err != nil {
			ws.log.Errorf("%v  Error: %v", ws.conn.RemoteAddr(), err)
			ws.client.Close(err)
			ws.client = nil
			return
		}

		ws.lastActive = time.Now()
		t := strings.ToLower(msg.Type)
		if t == "ping" {
			ws.SendResponse("pong", nil)
			continue
		}

		ws.log.Debugf("%v  Message: %+v", ws.conn.RemoteAddr(), msg)
		handled := false
		for _, handler := range ws.handlers {
			if t == handler.t {
				handled = true
				if err := handler.f(msg); err != nil {
					ws.log.Errorf("%v  Type %v  Error %v", ws.conn.RemoteAddr(), t, err)
					ws.client.Close(err)
					ws.client = nil
					return
				}
				break
			}
		}
		if !handled {
			ws.log.Errorf("%v  Type %q not handled.  Data %+v", ws.conn.RemoteAddr(), msg.Type, msg.Data.JSON(false))
		}
	}
}
