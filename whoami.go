package githubcliapp

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

func Whoami() string {
	var name struct {
		Login string `json:"login"`
	}
	_, err := os.Stat("./.token")
	if os.IsNotExist(err) {
		log.Fatal("You are not authorized. Run the `login` command.")
	}

	token, err := os.ReadFile("./.token")
	Check(err)

	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	Check(err)

	req.Header.Set("User-Agent", "cli-github-application")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+string(token))

	client := &http.Client{}
	resp, err := client.Do(req)
	Check(err)

	body, err := io.ReadAll(resp.Body)
	Check(err)

	err = json.Unmarshal(body, &name)
	Check(err)

	return name.Login
}
