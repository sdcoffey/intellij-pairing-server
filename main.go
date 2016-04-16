package main

import (
	"github.braintreeps.com/braintree/intellipair-server/sockethelper"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/{sessionId}", serveWs)
	http.ListenAndServe(":4000", router)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var sessions = make(map[string]*sockethelper.WebsocketSession)

func serveWs(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	sessionId := vars["sessionId"]
	if sessionId == "" {
		http.Error(writer, "Invalid sessionid", 400)
	}

	ws, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		panic(err)
		http.Error(writer, err.Error(), 400)
		return
	}

	if session, ok := sessions[sessionId]; ok {
		session.Register <- sockethelper.NewWebsocketReader(ws, session)
	} else {
		session := sockethelper.NewWebsocketSession(sessionId)
		session.Register <- sockethelper.NewWebsocketReader(ws, session)
		sessions[sessionId] = session
	}
}
