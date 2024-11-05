package stats

import (
    "context"
    "log"
    "sort"
    "sync"
    "time"
    "strings"
    "reddit-stats/internal/api"
)

type Stats struct {
    TopPosts       []api.Post
    UserPostCounts map[string]int
    Subreddit     string
}

type Collector struct {
    client     *api.RedditClient
    mu         sync.RWMutex
    stats      map[string]Stats
    seenPosts  map[string]map[string]bool
    subreddits []string
}

func NewCollector(client *api.RedditClient) *Collector {
    subreddits := strings.Split(client.Config.Subreddit, ",")
    statsMap := make(map[string]Stats)
    seenPosts := make(map[string]map[string]bool)
    
    for _, sub := range subreddits {
        sub = strings.TrimSpace(sub)
        if sub != "" {
            statsMap[sub] = Stats{
                TopPosts:       make([]api.Post, 0),
                UserPostCounts: make(map[string]int),
                Subreddit:     sub,
            }
            seenPosts[sub] = make(map[string]bool)
        }
    }
    
    return &Collector{
        client:     client,
        stats:      statsMap,
        seenPosts:  seenPosts,
        subreddits: subreddits,
    }
}

func (c *Collector) Start(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            for _, subreddit := range c.subreddits {
                subreddit = strings.TrimSpace(subreddit)
                if subreddit == "" {
                    continue
                }
                
                posts, err := c.client.GetPosts(subreddit)
                if err != nil {
                    log.Printf("Error fetching posts for %s: %v", subreddit, err)
                    continue
                }
                c.processPosts(subreddit, posts)
            }
            time.Sleep(10 * time.Second)
        }
    }
}

func (c *Collector) processPosts(subreddit string, posts []api.Post) {
    c.mu.Lock()
    defer c.mu.Unlock()

    stats := c.stats[subreddit]
    for _, post := range posts {
        if !c.seenPosts[subreddit][post.ID] {
            c.seenPosts[subreddit][post.ID] = true
            stats.UserPostCounts[post.Author]++
            stats.TopPosts = append(stats.TopPosts, post)
        }
    }

    sort.Slice(stats.TopPosts, func(i, j int) bool {
        return stats.TopPosts[i].Upvotes > stats.TopPosts[j].Upvotes
    })

    if len(stats.TopPosts) > 100 {
        stats.TopPosts = stats.TopPosts[:100]
    }

    c.stats[subreddit] = stats
}

func (c *Collector) Stats() []Stats {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    result := make([]Stats, 0, len(c.subreddits))
    for _, subreddit := range c.subreddits {
        subreddit = strings.TrimSpace(subreddit)
        if subreddit == "" {
            continue
        }
        if stats, ok := c.stats[subreddit]; ok {
            result = append(result, stats)
        }
    }
    return result
}

type Reporter struct {
    collector *Collector
    interval  time.Duration
}

func NewReporter() *Reporter {
    return &Reporter{
        interval: 15 * time.Second,
    }
}

func (r *Reporter) Report(allStats []Stats) {
    for _, stats := range allStats {
        log.Printf("\n=== Stats for r/%s ===\n", stats.Subreddit)
        
        log.Println("\n=== Top Posts ===")
        for i, post := range stats.TopPosts {
            if i >= 5 {
                break
            }
            log.Printf("%d. %s by %s (%d upvotes)\n", i+1, post.Title, post.Author, post.Upvotes)
        }

    log.Println("\n=== Most Active Users ===")
    type userCount struct {
        user  string
        count int
    }
    users := make([]userCount, 0, len(stats.UserPostCounts))
    for user, count := range stats.UserPostCounts {
        users = append(users, userCount{user, count})
    }

    sort.Slice(users, func(i, j int) bool {
        return users[i].count > users[j].count
    })

    for i, uc := range users {
        if i >= 5 {
            break
            }
            log.Printf("%d. %s (%d posts)\n", i+1, uc.user, uc.count)
        }
    }
}


