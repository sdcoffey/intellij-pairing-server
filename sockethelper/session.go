package sockethelper

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
						session.Unregister <- guest
					}
				}
			}
		}
	}
}
