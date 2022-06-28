package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Server struct {
	ListenPort     string
	LaravelAuthURL string
	Hub            *Hub
	ApiKey         string
}

// NewServer creates a new Server
func NewServer(listenPort string, laravelAuthURL string, apiKey string) *Server {
	return &Server{
		ListenPort:     listenPort,
		LaravelAuthURL: laravelAuthURL,
		Hub:            NewHub(),
		ApiKey:         apiKey,
	}
}

// serveSSE manages the SSE connection
func (srv *Server) serveSSE(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	log.Println("client connected")

	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", getAllowedOrigin(req))
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	channelName := req.FormValue("channel_name")
	sid := req.FormValue("sid")

	// call Laravel to check authorization
	data := url.Values{}
	data.Set("channel_name", channelName)

	httpClient := &http.Client{}
	larReq, err := http.NewRequest("POST", srv.LaravelAuthURL, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("ERROR: NewRequest: %s", err)
		return
	}

	larReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// let's replicate cookies
	for _, cookie := range req.Cookies() {
		larReq.AddCookie(cookie)
	}

	larRes, err := httpClient.Do(larReq)
	if err != nil {
		log.Printf("ERROR: httpClient.Do: %s", err)
		return
	}
	defer larRes.Body.Close()

	fmt.Println("Laravel response code:", larRes.StatusCode)

	authorized := false

	if larRes.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(larRes.Body)
		if err != nil {
			log.Printf("ERROR: ReadAll: %s", err)
			return
		}
		bodyString := string(bodyBytes)

		if bodyString == "true" {
			authorized = true
		}
	}

	if !authorized {
		log.Println("client unauthorized")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	// return a new session cookie with SID to identify future client actions
	sidCookie := &http.Cookie{
		Name:  "io",
		Value: sid,
		// Domain: "localhost", // read this in env ?
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, sidCookie)
	flusher.Flush()

	log.Printf("%s authenticated for: %s", sid, channelName)

	client := srv.Hub.Register(sid, channelName)

	// send a first event (some setup needs this to keep the connection alive)
	fmt.Fprintf(w, "event: ping\n")
	fmt.Fprintf(w, "data: {}\n")
	fmt.Fprintf(w, "\n\n")
	flusher.Flush()

	done := make(chan bool)
	go func() {
		for {
			select {
			/*case <-time.After(1 * time.Second):
			log.Println("ping")*/
			case event := <-client.ch:
				json, _ := json.Marshal(event.Payload)

				fmt.Fprintf(w, "event: %s\n", event.ChannelName)
				fmt.Fprintf(w, "data: %s\n", json)
				fmt.Fprintf(w, "\n\n")
				flusher.Flush()

			case <-req.Context().Done():
				done <- true
			}
		}
	}()
	<-done
	log.Println("client disconnected")
	srv.Hub.Unregister(client)
}

// serveEvent allow to broadcast an event
func (srv *Server) serveEvent(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	//Check API key
	if req.Header.Get("X-Api-Key") != srv.ApiKey {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
func (srv *Server) Run() {
	go srv.Hub.Run()

	http.HandleFunc("/sse", srv.serveSSE)
	http.HandleFunc("/event", srv.serveEvent)
	err := http.ListenAndServe(srv.ListenPort, nil)
	if err != nil {
		log.Fatal(err)
	}
}
