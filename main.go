package main

import (
	"flag"
	"fmt"
)

const Version = "0.0.1"

var ListenPort string

func main() {
	port := flag.Int("p", 8888, "listening port")
	version := flag.Bool("v", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(Version)
		return
	}

	ListenPort = fmt.Sprintf("localhost:%d", *port)

	fmt.Printf("listening on http://%s/\n", ListenPort)
	StartServer()
}
