package stats

import (
	"context"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
	"reddit-stats/internal/api"
)

type Stats struct {
	TopPosts       []api.Post
	UserPostCounts map[string]int
}

type Collector struct {
	client    *api.RedditClient
	mu        sync.RWMutex
	stats     Stats
	seenPosts map[string]bool
}

func NewCollector(client *api.RedditClient) *Collector {
	return &Collector{
		client: client,
		stats: Stats{
			TopPosts:       make([]api.Post, 0),
			UserPostCounts: make(map[string]int),
		},
		seenPosts: make(map[string]bool),
	}
}

func (c *Collector) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	userCounts := make(map[string]int)
	for k, v := range c.stats.UserPostCounts {
		if v > 0 {
			userCounts[k] = v
		}
	}

	posts := make([]api.Post, len(c.stats.TopPosts))
	copy(posts, c.stats.TopPosts)

	return Stats{
		TopPosts:       posts,
		UserPostCounts: userCounts,
	}
}

func (c *Collector) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	log.Println("Starting collector...")

	if err := c.update(); err != nil {
		log.Printf("Initial update error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.update(); err != nil {
				log.Printf("Update error: %v", err)
				continue
			}
		}
	}
}

func (c *Collector) update() error {
	posts, err := c.client.GetPosts(c.client.Config.Subreddit)
	if err != nil {
		return err
	}

	log.Printf("Fetched %d posts", len(posts))

	c.mu.Lock()
	defer c.mu.Unlock()

	var updatedPosts []api.Post
	postMap := make(map[string]api.Post)

	for _, post := range c.stats.TopPosts {
		postMap[post.ID] = post
	}

	for _, post := range posts {
		postMap[post.ID] = post
		
		if !c.seenPosts[post.ID] {
			c.seenPosts[post.ID] = true
			c.stats.UserPostCounts[post.Author]++
		}
	}

	for _, post := range postMap {
		updatedPosts = append(updatedPosts, post)
	}

	sort.Slice(updatedPosts, func(i, j int) bool {
		return updatedPosts[i].Upvotes > updatedPosts[j].Upvotes
	})

	if len(updatedPosts) > 100 {
		updatedPosts = updatedPosts[:100]
	}

	c.stats.TopPosts = updatedPosts

	return nil
}

type Reporter struct {
	reportInterval time.Duration
}

func NewReporter() *Reporter {
	return &Reporter{
		reportInterval: 15 * time.Second,
	}
}

func (r *Reporter) Report(stats Stats) {
	log.Println("\n=== Top Posts ===")
	for i, post := range stats.TopPosts {
		if i >= 5 {
			break
		}
		log.Printf("%d. %s by %s (%d upvotes)\n", i+1, post.Title, post.Author, post.Upvotes)
	}

	log.Println("\n=== Most Active Users ===")
	users := make([]struct {
		name  string
		count int
	}, 0, len(stats.UserPostCounts))

	for name, count := range stats.UserPostCounts {
		if count > 0 {
			users = append(users, struct {
				name  string
				count int
			}{name, count})
		}
	}

	sort.Slice(users, func(i, j int) bool {
		if users[i].count == users[j].count {
			return users[i].name < users[j].name
		}
		return users[i].count > users[j].count
	})

	limit := 5
	if len(users) < limit {
		limit = len(users)
	}

	for i := 0; i < limit; i++ {
		log.Printf("%d. %s (%d posts)\n", i+1, users[i].name, users[i].count)
	}

	log.Println(strings.Repeat("-", 50))
}
