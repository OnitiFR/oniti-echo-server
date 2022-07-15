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

type ServerSSE struct {
	ListenPort      string
	LaravelAuthURL  string
	Hub             *Hub
	AllowedOrigins  []string
	SIDCookieDomain string
}

// NewServerSSE creates a new server for SSE
func NewServerSSE(listenPort string, hub *Hub, laravelAuthURL string, allowedOrigins []string, sidCookieDomain string) *ServerSSE {
	return &ServerSSE{
		ListenPort:      listenPort,
		LaravelAuthURL:  laravelAuthURL,
		Hub:             hub,
		AllowedOrigins:  allowedOrigins,
		SIDCookieDomain: sidCookieDomain,
	}
}

// serveSSE manages the SSE connection
func (srv *ServerSSE) serveSSE(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("client connected")

	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	origin := req.Header.Get("Origin")
	found := false
	for _, allowedOrigin := range srv.AllowedOrigins {
		if origin == allowedOrigin {
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Origin not allowed", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", origin)

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
		Name:     "io",
		Value:    sid,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	if srv.SIDCookieDomain != "" {
		sidCookie.Domain = srv.SIDCookieDomain
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
				json, err := json.Marshal(event.Payload)

				if err != nil {
					log.Printf("ERROR: json.Marshal: %s", err)
					return
				}

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

// Run the server
func (srv *ServerSSE) Run() {
	fmt.Printf("SSE server listening on port %s\n", srv.ListenPort)
	for _, allowedOrigin := range srv.AllowedOrigins {
		fmt.Printf("Allowed origin: %s\n", allowedOrigin)
	}

	http.HandleFunc("/sse", srv.serveSSE)
	err := http.ListenAndServe(srv.ListenPort, nil)
	if err != nil {
		log.Fatal(err)
	}
}
