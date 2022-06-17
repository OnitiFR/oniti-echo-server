package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func serveRoot(w http.ResponseWriter, req *http.Request) {
	str := `
	<!DOCTYPE html>
	<html>
		<head>
			<title>SSE</title>
		</head>
		<body>
			<h1>SSE</h1>
			<script>
				// créer un sid
				var source = new EventSource("/sse?chan=private-collection-crud-roles&sid=AAAG", { withCredentials: true })
				source.addEventListener("ping", (ev) => { console.log(ev) })
			</script>
		</body>
	</html>
	`

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write([]byte(str))
}

func serveSSE(w http.ResponseWriter, req *http.Request) {
	log.Println("client connected")

	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	chanName := req.FormValue("chan")

	fmt.Println("chanName:", chanName)

	// on contacte Laravel pour vérifier les droits
	// POST/GET/HEAD http://localhost:8000/broadcasting/auth
	// duplicate all cookies
	/*for _, cookie := range req.Cookies() {
		//w.Header().Add("Set-Cookie", cookie.String())
	}*/
	// associer une clée propre à la connexion pour le toOthers

	authorized := true

	if !authorized {
		log.Println("client unauthorized")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	// renvoyer un nouveau cookie de session SID

	// on "subscribe" au channel

	done := make(chan bool)
	go func() {
		for {
			select {
			// on écoute le channel
			case <-time.After(1 * time.Second):
				log.Println("ping")
				fmt.Fprintf(w, "event: ping\n")
				fmt.Fprintf(w, "data: {val: 12}\n")
				fmt.Fprintf(w, "\n\n")
				flusher.Flush()

			case <-req.Context().Done():
				done <- true
			}
		}
	}()
	<-done
	log.Println("client disconnected")
}

// serveEvent permet à Laravel d'envoyer un nouvel event
func serveEvent(w http.ResponseWriter, req *http.Request) {
	// parcourir les subscriptions et envoyer un event SAUF
	// pour le SID de l'émetteur
}

func StartServer() {
	http.HandleFunc("/", serveRoot)
	http.HandleFunc("/sse", serveSSE)
	http.HandleFunc("/event", serveEvent)
	err := http.ListenAndServe(ListenPort, nil)
	if err != nil {
		log.Fatal(err)
	}
}
