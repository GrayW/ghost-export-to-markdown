package ghost

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Client handles all Ghost API interactions
type Client struct {
	apiURL string
	apiKey string
	http   *http.Client
}

// NewClient creates a new Ghost API client
func NewClient(apiURL, apiKey string) *Client {
	return &Client{
		apiURL: strings.TrimPrefix(apiURL, "https://"),
		apiKey: apiKey,
		http:   &http.Client{},
	}
}

// FetchPosts retrieves all posts from the Ghost API, using pagination
func (c *Client) FetchPosts() ([]Post, error) {
	var allPosts []Post
	page := 1
	limit := 100 // Max allowed by Ghost API

	for {
		url := fmt.Sprintf("https://%s/ghost/api/content/posts/?key=%s&page=%d&limit=%d", c.apiURL, c.apiKey, page, limit)

		resp, err := c.http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch posts: %w", err)
		}
		defer resp.Body.Close()

		var result struct {
			Posts  []Post     `json:"posts,omitempty"`
			Meta   PostMeta   `json:"meta,omitempty"`
			Errors []APIError `json:"errors,omitempty"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		if len(result.Errors) > 0 {
			return nil, fmt.Errorf("API error: %s (code: %s)",
					       result.Errors[0].Message,
			  result.Errors[0].Code,
			)
		}

		if len(result.Posts) == 0 {
			break
		}

		allPosts = append(allPosts, result.Posts...)
		page++
	}

	return allPosts, nil
}

// PostMeta contains pagination info from the Ghost API
type PostMeta struct {
	Pagination struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
		Pages int `json:"pages"`
		Total int `json:"total"`
		Next  int `json:"next,omitempty"`
		Prev  int `json:"prev,omitempty"`
	} `json:"pagination"`
}
