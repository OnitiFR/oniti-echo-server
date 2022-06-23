package main

// Event (from Hub point of vue)
type Event struct {
	ChannelName string
	Socket      string
	Payload     Payload
}

// Payload (event from SSE point of vue)
type Payload struct {
	Data  string
	Event string
}
