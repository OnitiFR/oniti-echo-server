package main

import (
	"fmt"
	"log"
	"net/http"
)

type ServerEvent struct {
	ListenPort string
	Hub        *Hub
}

// NewServerEvent creates a new server for events
func NewServerEvent(listenPort string, hub *Hub) *ServerEvent {
	return &ServerEvent{
		ListenPort: listenPort,
		Hub:        hub,
	}
}

// serveEvent allow to broadcast an event
func (srv *ServerEvent) serveEvent(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fmt.Printf("%+v\n", req.FormValue("ChannelName"))
	ev := &Event{
		ChannelName: req.Form.Get("ChannelName"),
		Socket:      req.Form.Get("socket"),
		Payload: Payload{
			Data:  req.Form.Get("payload"),
			Event: req.Form.Get("event"),
		},
	}
	srv.Hub.Publish(ev)

}

// Run the server
func (srv *ServerEvent) Run() {
	fmt.Printf("ServerEvent listening on %s\n", srv.ListenPort)
	http.HandleFunc("/event", srv.serveEvent)
	err := http.ListenAndServe(srv.ListenPort, nil)
	if err != nil {
		log.Fatal(err)
	}
}
