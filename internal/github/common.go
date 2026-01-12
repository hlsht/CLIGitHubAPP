package github

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func Check(err error) {
	if err != nil {
		log.Fatalf("%s: %s", os.Args[0], err)
	}
}

func (a *App) getAuthToken() string {
	token, err := os.ReadFile(a.tokenPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error reading token file: %s\n", err)
		os.Exit(1)
	}
	return strings.TrimSpace(string(token))
}
