package remarkable

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

const (
	DefaultAuthHost = "https://webapp-prod.cloud.remarkable.engineering"
	DefaultSyncHost = "https://eu.tectonic.remarkable.com"
)

type Client struct {
	authHost   string
	syncHost   string
	httpClient *http.Client
}

func NewClient(authHost string) *Client {
	return &Client{
		authHost:   authHost,
		syncHost:   DefaultSyncHost,
		httpClient: &http.Client{},
	}
}

func (c *Client) RegisterDevice(code string) (string, error) {
	body := map[string]string{
		"code":       code,
		"deviceDesc": "desktop-macos",
		"deviceID":   uuid.New().String(),
	}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Post(
		c.authHost+"/token/json/2/device/new",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registration failed: status %d", resp.StatusCode)
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(token), nil
}

func (c *Client) RefreshUserToken(deviceToken string) (string, error) {
	req, err := http.NewRequest("POST", c.authHost+"/token/json/2/user/new", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+deviceToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token refresh failed: status %d", resp.StatusCode)
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(token), nil
}
