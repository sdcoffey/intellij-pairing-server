package sockethelper

import (
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

type WebsocketMessage struct {
	id, message string
}

type WebsocketReader struct {
	memberId string
	socket   *websocket.Conn
	Send     chan string
	Session  *WebsocketSession
}

func NewWebsocketReader(socket *websocket.Conn, session *WebsocketSession) *WebsocketReader {
	wsr := &WebsocketReader{
		uuid.NewV4().String(),
		socket,
		make(chan string, 100),
		session,
	}
	go wsr.run()
	return wsr
}

func (wsr *WebsocketReader) Close() {
	wsr.socket.Close()
	close(wsr.Send)
}

func (wsr *WebsocketReader) run() {
	go wsr.checkEvents()
	for {
		_, dat, err := wsr.socket.ReadMessage()
		if err == nil {
			wsr.Session.Broadcast <- WebsocketMessage{wsr.memberId, string(dat)}
		}
	}
}

func (wsr *WebsocketReader) checkEvents() {
	for newMessage := range wsr.Send {
		wsr.socket.WriteMessage(websocket.TextMessage, []byte(newMessage))
	}
}
