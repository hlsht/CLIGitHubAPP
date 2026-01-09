package githubcliapp

import (
	"log"
	"os"
	"strings"
)

func Check(err error) {
	if err != nil {
		log.Fatalf("%s: %s", os.Args[0], err)
	}
}

func GetAuthToken() string {
	token, err := os.ReadFile("./.token")
	Check(err)
	return strings.TrimSpace(string(token))
}
