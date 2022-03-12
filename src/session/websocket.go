package session

import (
	"compress/flate"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"me.game/src/bundle"
	"me.game/src/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    utils.BufferSize,
	WriteBufferSize:   utils.BufferSize,
	HandshakeTimeout:  1 * time.Second,
	EnableCompression: true,
}

type websocketProt struct {
	ws *websocket.Conn
}

func NewWebSocket(w http.ResponseWriter, r *http.Request) (SessionHandle, error) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	ws.SetReadLimit(utils.BufferSize)
	ws.SetCompressionLevel(flate.BestSpeed)

	return &websocketProt{
		ws: ws,
	}, nil
}

func (ws *websocketProt) Close() error {
	return ws.ws.Close()
}

func (ws *websocketProt) Read(b *bundle.Broadcast) error {
	return ws.ws.ReadJSON(b)
}

func (ws *websocketProt) ReadName() (string, error) {
	ws.ws.SetReadDeadline(time.Now().Add(utils.NameTimeout))

	var object json.RawMessage
	b := bundle.Broadcast{
		Message: &object,
	}

	err := ws.ws.ReadJSON(&b)
	if err != nil {
		return "", err
	}

	if b.Code == bundle.AddReqCode {
		var add bundle.AddRequest
		err = json.Unmarshal(object, &add)
		if err != nil {
			return "", err
		}

		ws.ws.SetReadDeadline(time.Now().Add(utils.ReadTimeout))
		return add.Name, nil
	}

	return "", errorName
}

func (ws *websocketProt) RemoteAddr() string {
	return ws.ws.RemoteAddr().String()
}

func (ws *websocketProt) Write(b bundle.Broadcast) error {
	ws.ws.SetWriteDeadline(time.Now().Add(utils.WriteTimeout))
	return ws.ws.WriteJSON(b)
}
