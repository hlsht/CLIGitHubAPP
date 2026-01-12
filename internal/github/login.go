package github

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type AccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ClientId     string `json:"client_id"`
	DeviceCode   string `json:"device_code"`
	GrantType    string `json:"grant_type"`
	RepotisoryId string `json:"repository_id"`
	ErrorResp    string `json:"error"`
}

const ClientID = "Iv23lisaqWSy9gCaljlM"

func (a *App) requestDeviceCode() ([]byte, error) {
	data := url.Values{}
	data.Set("client_id", ClientID)

	req, err := a.NewRequest("POST", "https://github.com/login/device/code",
		strings.NewReader(data.Encode()),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"})
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error while requesting: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ghcli: bad request status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (a *App) requestToken(deviceCode string) ([]byte, error) {
	data := url.Values{}
	data.Set("client_id", ClientID)
	data.Set("device_code", deviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	req, err := a.NewRequest("POST", "https://github.com/login/oauth/access_token",
		strings.NewReader(data.Encode()),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"})
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error while requesting: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ghcli: bad request status code: %d\n", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (a *App) pollForToken(deviceCode string, interval int) {
	for {
		var at AccessTokenResponse
		resp, err := a.requestToken(deviceCode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ghcli: error while requesting for token: %s\n", err)
			os.Exit(1)
		}

		if err := json.Unmarshal(resp, &at); err != nil {
			log.Fatalf("ghcli: error unmarshaling json")
		}

		if at.ErrorResp != "" {
			switch at.ErrorResp {
			case "authorization_pending":
				time.Sleep(time.Duration(interval) * time.Second)
				continue
			case "slow_down":
				time.Sleep(time.Duration(interval+5) * time.Second)
				continue
			case "expired_token":
				fmt.Println("The device code has expired. Please run `login` again.")
				os.Exit(1)
			case "access_denied":
				fmt.Println("Login cancelled by user.")
				os.Exit(1)
			}
		}

		tokenDir := filepath.Join(os.Getenv("HOME"), ".config", "ghcli")
		if err := os.MkdirAll(tokenDir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "ghcli: error creating directory %s: %s\n", tokenDir, err)
			os.Exit(1)
		}
		if err := os.WriteFile(a.tokenPath, []byte(at.AccessToken), 0o600); err != nil {
			fmt.Fprintf(os.Stderr, "ghcli: error writing to file %s: %s\n", a.tokenPath, err)
			os.Exit(1)
		}
		break
	}
}

func (a *App) Login() {
	var dc DeviceCodeResponse
	resp, err := a.requestDeviceCode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghcli: error requesting device code: %s\n", err)
		os.Exit(1)
	}

	if err := json.Unmarshal(resp, &dc); err != nil {
		log.Fatalf("ghcli: error unmarshaling json\n")
		os.Exit(1)
	}

	fmt.Printf("Please visit: %s\n", dc.VerificationURI)
	fmt.Printf("and enter code: %s\n", dc.UserCode)

	a.pollForToken(dc.DeviceCode, dc.Interval)

	fmt.Println("Successfully authenticated!")
}
