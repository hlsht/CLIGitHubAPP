package githubcliapp

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

type Repo struct {
	Name string `json:"name"`
}

func ListRepos() []Repo {
	var repos []Repo

	req, err := http.NewRequest("GET", "https://api.github.com/user/repos", nil)
	Check(err)

	token, err := os.ReadFile("./.token")
	Check(err)

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+string(token))

	client := &http.Client{}
	resp, err := client.Do(req)
	Check(err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	Check(err)

	err = json.Unmarshal(body, &repos)
	Check(err)

	return repos
}
