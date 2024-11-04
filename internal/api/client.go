package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"reddit-stats/internal/config"
	"reddit-stats/internal/ratelimit"
)

type RedditClient struct {
	Config      *config.Config
	httpClient  *http.Client
	limiter     *ratelimit.RateLimiter
	accessToken string
}

func NewRedditClient(cfg *config.Config) (*RedditClient, error) {
	client := &RedditClient{
		Config: cfg,
		httpClient: &http.Client{
			Timeout: time.Second * 30,
		},
		limiter: ratelimit.NewRateLimiter(),
	}

	if err := client.refreshToken(); err != nil {
		return nil, fmt.Errorf("failed to get initial token: %w", err)
	}

	return client, nil
}

func (c *RedditClient) GetPosts(subreddit string) ([]Post, error) {
	url := fmt.Sprintf("https://oauth.reddit.com/r/%s/new.json?limit=100", subreddit)

	if err := c.limiter.Wait(); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("User-Agent", "reddit-stats/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	c.limiter.UpdateFromHeaders(resp.Header)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	var listing ListingResponse
	if err := json.NewDecoder(resp.Body).Decode(&listing); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return listing.Posts(), nil
}

func (c *RedditClient) refreshToken() error {
	tokenURL := "https://www.reddit.com/api/v1/access_token"

	req, err := http.NewRequest("POST", tokenURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.SetBasicAuth(c.Config.ClientID, c.Config.ClientSecret)
	req.Header.Set("User-Agent", "reddit-stats/1.0")

	q := req.URL.Query()
	q.Add("grant_type", "client_credentials")
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 status code from token endpoint: %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	return nil
}
