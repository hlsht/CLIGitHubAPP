package githubcliapp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
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

func requestDeviceCode() ([]byte, error) {
	data := url.Values{}
	data.Set("client_id", ClientID)

	req, err := http.NewRequest("POST", "https://github.com/login/device/code", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	Check(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: bad request status code: %d", os.Args[0], resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func requestToken(deviceCode string) ([]byte, error) {
	data := url.Values{}
	data.Set("client_id", ClientID)
	data.Set("device_code", deviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	Check(err)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: bad request status code: %d", os.Args[0], resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func parseResponse(resp *http.Response) (*DeviceCodeResponse, error) {
	var result DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("%s: error unmarshaling json", os.Args[0])
	}
	return &result, nil
}

func pollForToken(deviceCode string, interval int) {
	for {
		var at AccessTokenResponse
		resp, err := requestToken(deviceCode)
		Check(err)

		if err != nil {
			log.Fatalf("%s: error unmarshaling json", os.Args[0])
		}

		if err := json.Unmarshal(resp, &at); err != nil {
			log.Fatalf("%s: error unmarshaling json", os.Args[0])
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

		tokenName := "./.token"
		tokenFile, err := os.Create(tokenName)
		Check(err)
		defer tokenFile.Close()
		out := bufio.NewWriter(tokenFile)

		_, err = out.WriteString(at.AccessToken)
		Check(err)
		out.Flush()
		break
	}
}

func Login() {
	var dc DeviceCodeResponse
	resp, err := requestDeviceCode()
	Check(err)

	if err := json.Unmarshal(resp, &dc); err != nil {
		log.Fatalf("%s: error unmarshaling json", os.Args[0])
	}

	fmt.Printf("Please visit: %s\n", dc.VerificationURI)
	fmt.Printf("and enter code: %s\n", dc.UserCode)

	pollForToken(dc.DeviceCode, dc.Interval)

	fmt.Println("Successfully authenticated!")
}
