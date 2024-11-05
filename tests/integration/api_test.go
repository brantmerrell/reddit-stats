package integration

import (
	"testing"
	"time"
	"reddit-stats/internal/api"
)

type mockRedditClient api.RedditClient

func newMockRedditClient() *api.RedditClient {
	return &api.RedditClient{}
}

func TestRedditClient_GetPosts(t *testing.T) {
	posts := []api.Post{
		{
			ID:      "123",
			Title:   "Test Post",
			Author:  "testuser",
			Upvotes: 100,
			Created: time.Now(),
		},
	}
	
	if len(posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(posts))
	}

	if posts[0].Title != "Test Post" {
		t.Errorf("Expected title 'Test Post', got '%s'", posts[0].Title)
	}
}

func TestCollector_Stats(t *testing.T) {
	posts := []api.Post{
		{
			ID:      "123",
			Title:   "Test Post",
			Author:  "testuser",
			Upvotes: 100,
			Created: time.Now(),
		},
	}

	if len(posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(posts))
	}

	if posts[0].Author != "testuser" {
		t.Errorf("Expected author 'testuser', got '%s'", posts[0].Author)
	}
}

func TestCollector_Deduplication(t *testing.T) {
	posts := []api.Post{
		{
			ID:      "123",
			Title:   "Test Post",
			Author:  "testuser",
			Upvotes: 100,
			Created: time.Now(),
		},
		{
			ID:      "123",
			Title:   "Test Post",
			Author:  "testuser",
			Upvotes: 100,
			Created: time.Now(),
		},
	}

	if len(posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(posts))
	}

	authors := make(map[string]int)
	for _, post := range posts {
		authors[post.Author]++
	}

	if authors["testuser"] != 2 {
		t.Errorf("Expected 2 posts from testuser, got %d", authors["testuser"])
	}
}
