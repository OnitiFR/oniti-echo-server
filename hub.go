package main

import "fmt"

// Hub centralizes all clients and allow to publish events
type Hub struct {
	clients    map[string]*HubClient
	register   chan *HubClient
	unregister chan *HubClient
	events     chan *Event
}

// HubClient is a client of the Hub
type HubClient struct {
	key         string
	sid         string
	channelName string
	ch          chan *Event
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*HubClient),
		events:     make(chan *Event),
		register:   make(chan *HubClient),
		unregister: make(chan *HubClient),
	}
}

// Run will start the Hub, allowing messages to be sent and received
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.key] = client
			fmt.Printf("new client: %s\n", client.key)
		case client := <-h.unregister:
			fmt.Printf("del client: %s\n", client.key)
			if _, ok := h.clients[client.key]; ok {
				delete(h.clients, client.key)
				close(client.ch)
			} else {
				fmt.Printf("client %s not found\n", client.key)
			}
		case event := <-h.events:
			fmt.Printf("on %s…\n", event.ChannelName)
			for _, client := range h.clients {
				if client.channelName == event.ChannelName {
					// TODO: if not self
					if client.sid != event.Socket {
						fmt.Printf("…to client %s\n", client.key)
						client.ch <- event
					}
				}
			}
			fmt.Printf("event end.\n")
		}
	}
}

// Register a new client of the Hub
func (h *Hub) Register(sid string, channelName string) *HubClient {
	client := &HubClient{
		key:         sid + "-" + channelName,
		ch:          make(chan *Event),
		sid:         sid,
		channelName: channelName,
	}
	h.register <- client
	return client
}

// Unregister a client of the Hub
func (h *Hub) Unregister(client *HubClient) {
	h.unregister <- client
}

// Publish an event to the Hub
func (h *Hub) Publish(event *Event) {
	h.events <- event
}
