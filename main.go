package main

import (
	"flag"
	"fmt"
)

const Version = "0.0.1"

func main() {
	port := flag.Int("p", 8888, "listening port")
	version := flag.Bool("v", false, "show version")
	apiKey := flag.String("k", "", "api key")

	flag.Parse()

	if *version {
		fmt.Println(Version)
		return
	}

	// Listen on all interfaces
	listenPort := fmt.Sprintf(":%d", *port)

	fmt.Printf("listening on http://%s/\n", listenPort)

	srv := NewServer(listenPort, "http://localhost:8000/broadcasting/auth", getenv("ONITI_ECHO_SERVER_API_KEY", *apiKey))
	srv.Run()
}
