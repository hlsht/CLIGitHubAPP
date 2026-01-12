package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func (a *App) Whoami() string {
	var name struct {
		Login string `json:"login"`
	}

	authToken := a.getAuthToken()
	req, err := a.NewRequest("GET", "https://api.github.com/user", nil, map[string]string{
		"Authorization": "Bearer " + authToken,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error making request: %s\n", err)
		os.Exit(1)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error while requesting: %s\n", err)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "ghcli: bad response status code: %s\n", resp.Status)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error reading response body: %s\n", err)
		os.Exit(1)
	}

	if err := json.Unmarshal(body, &name); err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error unmarshaling response: %s\n", err)
		os.Exit(1)
	}

	return name.Login
}
