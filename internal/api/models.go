package api

import "time"

type Post struct {
	ID      string
	Title   string
	Author  string
	Upvotes int
	Created time.Time
}

type ListingResponse struct {
	Data struct {
		Children []Child `json:"children"`
	} `json:"data"`
}

type Child struct {
	Data struct {
		ID      string  `json:"id"`
		Title   string  `json:"title"`
		Author  string  `json:"author"`
		Ups     int     `json:"ups"`
		Created float64 `json:"created_utc"`
	} `json:"data"`
}

func (l ListingResponse) Posts() []Post {
	posts := make([]Post, 0, len(l.Data.Children))
	for _, child := range l.Data.Children {
		posts = append(posts, Post{
			ID:      child.Data.ID,
			Title:   child.Data.Title,
			Author:  child.Data.Author,
			Upvotes: child.Data.Ups,
			Created: time.Unix(int64(child.Data.Created), 0),
		})
	}
	return posts
}
