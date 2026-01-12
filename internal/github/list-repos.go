package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Repo struct {
	Name string `json:"name"`
}

func (a *App) ListRepos() []Repo {
	var repos []Repo

	authToken := a.getAuthToken()
	req, err := a.NewRequest("GET", "https://api.github.com/user/repos", nil,
		map[string]string{
			"Authorization": "Bearer " + authToken,
		})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error making request: %s", err)
		os.Exit(1)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error while requesting: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "ghcli: bad response status code: %s\n", resp.Status)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error reading request body: %s\n", err)
		os.Exit(1)
	}

	err = json.Unmarshal(body, &repos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error unmarshaling json: %s\n", err)
		os.Exit(1)
	}

	return repos
}
