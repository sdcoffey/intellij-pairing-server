package main

import (
	"github.com/gorilla/websocket"
	"net/http"
)

func main() {
	http.HandleFunc("/testSession", serveWs)
	http.ListenAndServe(":4000", nil)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var sessions = make(map[string]*session)

func serveWs(writer http.ResponseWriter, request *http.Request) {
	// TODO: if sessionId path we already have, add the person as guest
	// TODO else create a new session with that person as the host

	ws, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		http.Error(writer, err.Error(), 400)
	} else {
		for {
			_, dat, err := ws.ReadMessage()
			if err != nil {
				println(err.Error())
				ws.Close()
				return
			}
			println(string(dat))
		}
	}
}

type session struct {
	host      *websocket.Conn
	guest     *websocket.Conn
	sessionId string
}

func (sess *session) run() {
	hostChannel := make(chan string)
	guestChannel := make(chan string)

	go func() {
		for sess.host != nil {
			_, dat, err := sess.host.ReadMessage()
			if err != nil {
				println(err.Error())
			} else if sess.guest != nil {
				guestChannel <- string(dat)
			}
		}
	}()

	go func() {
		for sess.guest != nil {
			_, dat, err := sess.guest.ReadMessage()
			if err != nil {
				println(err.Error())
			} else if sess.host != nil {
				hostChannel <- string(dat)
			}
		}
	}()

	for {
		select {
		case message := <-hostChannel:
			sess.host.WriteMessage(websocket.TextMessage, []byte(message))
		case message := <-guestChannel:
			sess.guest.WriteMessage(websocket.TextMessage, []byte(message))
		}
	}
}
