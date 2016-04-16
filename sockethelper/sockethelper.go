package sockethelper

import (
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

type WebsocketSession struct {
	sessionId  string
	Guests     map[*WebsocketReader]bool
	Register   chan *WebsocketReader
	Broadcast  chan WebsocketMessage
	Unregister chan *WebsocketReader
}

func NewWebsocketSession(sessionId string) *WebsocketSession {
	session := &WebsocketSession{
		sessionId:  sessionId,
		Guests:     make(map[*WebsocketReader]bool),
		Register:   make(chan *WebsocketReader),
		Unregister: make(chan *WebsocketReader),
		Broadcast:  make(chan WebsocketMessage),
	}
	go session.run()

	return session
}

func (session *WebsocketSession) run() {
	for {
		select {
		case guest := <-session.Register:
			session.Guests[guest] = true
		case guest := <-session.Unregister:
			delete(session.Guests, guest)
			guest.Close()
		case message := <-session.Broadcast:
			for guest, _ := range session.Guests {
				if message.id != guest.memberId {
					select {
					case guest.Send <- message.message:
					default:
						delete(session.Guests, guest)
						guest.Close()
					}
				}
			}
		}
	}
}

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
