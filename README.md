# Reddit Stats

This application tracks real-time statistics from Reddit subreddits. By default, it monitors [r/chess](https://www.reddit.com/r/chess/), but this can be configured to any subreddit. It collects data from Reddit's `/r/{subreddit}/new.json` endpoint, which is called every 10 seconds with respect to Reddit's rate limiting (600 requests per 10 minutes).

## Setup

**Setting up .env**:
```env
export REDDIT_CLIENT_ID=client_id
export REDDIT_CLIENT_SECRET=client_secret
export REDDIT_SUBREDDIT=chess
```

**installing dependencies**:
```bash
go mod download
```

**Building the app**:

```bash
go build -o reddit-stats ./cmd/tracker
```

**Running the app**:

```bash
./reddit-stats
```

**Viewing the output**:

```bash
tail -f reddit-stats.log
```
The application logs statistics every 15 seconds in this format:

```
2024/11/04 14:53:00 Starting collector...
2024/11/04 14:53:01 Fetched 100 posts
2024/11/04 14:53:11 Fetched 100 posts
2024/11/04 14:53:15 
=== Top Posts ===
2024/11/04 14:53:15 1. From my 1st grade son’s Chess “homework”. There is supposed to be a Mate in 1 here for white…am I going crazy? by andersonle09 (1723 upvotes)
2024/11/04 14:53:15 2. Showgirls play chess before a show, 1958 by EastEndChess (942 upvotes)
2024/11/04 14:53:15 3. Ding Liren in an interview on Take Take Take by Radiant-Increase-180 (889 upvotes)
2024/11/04 14:53:15 4. Viktor Korchnoi, a chess grandmaster who defected from the USSR, wore mirrored glasses during his world championship match against Soviet Anatoly Karpov to protect himself from "hypnotic attacks." He was convinced that the Soviets would try to cheat and then kill him. by 10mayyy (606 upvotes)
2024/11/04 14:53:15 5. Is bullet chess the reason why low-rated players aren’t making progress? by vggoi (448 upvotes)
2024/11/04 14:53:15 
=== Most Active Users ===
2024/11/04 14:53:15 1. ChessDrawsAreOK (2 posts)
2024/11/04 14:53:15 2. ExistingPrinciple137 (2 posts)
2024/11/04 14:53:15 3. Macbeth59 (2 posts)
2024/11/04 14:53:15 4. Necessary_Pattern850 (2 posts)
2024/11/04 14:53:15 5. Rhino887 (2 posts)
...
```

## Discussion

### Data Persistence

Currently, the application maintains statistics in memory. For production deployments requiring persistence, I would consider PostgreSQL, SQLite, or Redis depending on architectural considerations.

### Rate Limiting

The application implements sophisticated rate limiting that:
- Respects Reddit's API rate limits
- Uses response headers to adjust request timing
- Maintains a buffer to prevent limit exhaustion
- Distributes requests evenly across the rate limit window

### Design Patterns


The app has the following directory structure:

```text
.
├── README.md
├── cmd
│   └── tracker
│       └── main.go
├── go.mod
├── go.sum
├── internal
│   ├── api
│   │   ├── client.go
│   │   └── models.go
│   ├── config
│   │   └── config.go
│   ├── ratelimit
│   │   └── limiter.go
│   └── stats
│       └── stats.go
├── pkg
├── reddit-stats.log
└── tests
    └── integration
        └── api_test.go
```

It uses multiple goroutines to handle:
- Post collection from Reddit
- Statistics processing
- Periodic reporting
- Rate limit management

This allows for a clear separation between API client, statistics collection, and reporting. The statistics collector can be extended without modification, and new statistics can be added by implementing additional collectors. This exhibits the SOLID principles of architectural design.
