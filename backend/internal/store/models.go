package store

import "time"

type Post struct {
	ID        int       `json:"-"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Excerpt   string    `json:"excerpt,omitempty"`
	Tags      []string  `json:"tags"`
	Author    string    `json:"author"`
	WordCount int       `json:"wordCount"`
	ReadTime  string    `json:"readTime"`
	Published bool      `json:"published"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Project struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Lang        string `json:"lang"`
	Status      string `json:"status"`
	URL         string `json:"url"`
	SortOrder   int    `json:"sortOrder"`
}

type Admin struct {
	ID           int
	Username     string
	PasswordHash string
}
