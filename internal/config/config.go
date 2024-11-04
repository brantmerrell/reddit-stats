package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"time"
)

type Config struct {
	ClientID        string
	ClientSecret    string
	Subreddit       string
	UpdateInterval  time.Duration
	RateLimitBuffer float64
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: error loading .env file: %v\n", err)
	}

	config := &Config{
		ClientID:        os.Getenv("REDDIT_CLIENT_ID"),
		ClientSecret:    os.Getenv("REDDIT_CLIENT_SECRET"),
		Subreddit:       os.Getenv("REDDIT_SUBREDDIT"),
		UpdateInterval:  30 * time.Second,
		RateLimitBuffer: 0.9,
	}

	var missingVars []string
	if config.ClientID == "" {
		missingVars = append(missingVars, "REDDIT_CLIENT_ID")
	}
	if config.ClientSecret == "" {
		missingVars = append(missingVars, "REDDIT_CLIENT_SECRET")
	}
	if config.Subreddit == "" {
		missingVars = append(missingVars, "REDDIT_SUBREDDIT")
	}

	if len(missingVars) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missingVars)
	}

	return config, nil
}
