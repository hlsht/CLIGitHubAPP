package github

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type App struct {
	client    *http.Client
	login     string
	tokenPath string
}

func NewApp() *App {
	return &App{
		client:    &http.Client{Timeout: 10 * time.Second},
		tokenPath: filepath.Join(os.Getenv("HOME"), ".config", "ghcli", ".token"),
	}
}

func (a *App) NewRequest(method, url string, reader io.Reader, headers map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}

	// sets up default headers for all request from app
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "CLIGithubApp")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	// sets up the parameters passed to the function
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

func (a *App) CheckAuthorization() bool {
	authToken := a.getAuthToken()
	req, err := a.NewRequest("GET", "https://api.github.com/octocat", nil,
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

	return resp.StatusCode == http.StatusOK
}
