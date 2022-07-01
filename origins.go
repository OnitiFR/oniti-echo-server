package main

import (
	"os"
	"strings"
)

func GetAllowedOrigins() []string {
	domainsEnv := os.Getenv("_DOMAINS")

	if domainsEnv == "" {
		return []string{}
	}

	domains := strings.Split(domainsEnv, ",")

	for i, domain := range domains {
		domains[i] = "http://" + domain
		domains[i] = "https://" + domain
	}

	return domains
}
