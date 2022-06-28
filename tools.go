package main

import (
	"net/http"
	"os"
	"strings"
)

func getAllowedOrigin(req *http.Request) string {
	origin := req.Header.Get("Origin")

	domains := os.Getenv("_DOMAINS")
	domainsArray := strings.Split(domains, ",")

	for i, domain := range domainsArray {
		domainsArray[i] = "http://" + domain
		domainsArray[i] = "https://" + domain
	}

	allowedOrigins := []string{"http://localhost:8080"}
	allowedOrigins = append(allowedOrigins, domainsArray...)

	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return origin
		}
	}

	return ""
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
