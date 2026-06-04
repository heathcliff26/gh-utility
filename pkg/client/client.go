package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	DefaultEndpoint = "https://api.github.com"

	defaultClientTimeout = 2 * time.Second
)

type client struct {
	httpClient *http.Client
	endpoint   string
}

// Create a new client
func NewClient(endpoint string) *client {
	httpClient := &http.Client{
		Timeout: defaultClientTimeout,
	}

	return &client{
		httpClient: httpClient,
		endpoint:   endpoint,
	}
}

// Get an installations token for the GitHub app
// API endpoint: POST /app/installations/{installation_id}/access_tokens
func (c *client) GetToken(keyPath string, clientID string, installationID string) (string, error) {
	bytes, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read keyfile: %w", err)
	}

	// Need to replace literal `\n` with actual newlines.
	// Otherwise it is harder to use this in environments where the key might be passed as a single line string.
	keyStr := strings.ReplaceAll(string(bytes), `\n`, "\n")
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(keyStr))
	if err != nil {
		return "", fmt.Errorf("failed to parse keyfile: %w", err)
	}

	claims := &jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(9 * time.Minute)),
		Issuer:    clientID,
	}
	jwToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedJWToken, err := jwToken.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign JSONWebToken: %w", err)
	}

	req, err := newRequest(http.MethodPost, fmt.Sprintf("%s/app/installations/%s/access_tokens", c.endpoint, installationID), nil, signedJWToken)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("request returned non-201 status: %s", res.Status)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(res.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return tokenResp.Token, nil
}
