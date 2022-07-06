package main

import (
	"flag"
	"fmt"
)

const Version = "0.0.2"

func main() {
	portEvent := flag.Int("e", 8888, "listening port for events (localhost)")
	portSSE := flag.Int("s", 8889, "listening port for SSE (all interfaces)")
	allowedOrigin := flag.String("o", "http://localhost:8080", "default allowed origin, overridden by _DOMAINS env variable (comma separated list of *domains*)")
	authUrl := flag.String("a", "http://localhost:8000/broadcasting/auth", "auth url to contact")
	version := flag.Bool("v", false, "show version")

	flag.Parse()

	if *version {
		fmt.Println(Version)
		return
	}

	origins := GetAllowedOrigins()
	if len(origins) == 0 {
		origins = []string{*allowedOrigin}
	}

	listenPortEvent := fmt.Sprintf("localhost:%d", *portEvent)
	listenPortSSE := fmt.Sprintf(":%d", *portSSE)

	hub := NewHub()
	serveEvent := NewServerEvent(listenPortEvent, hub)
	serveSSE := NewServerSSE(listenPortSSE, hub, *authUrl, origins)

	go hub.Run()
	go serveEvent.Run()
	go serveSSE.Run()

	// wait forever
	select {}
}
