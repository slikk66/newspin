package news

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	apiKey  string
	baseUrl string
}

type Article struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`
	ImageUrl    string `json:"urlToImage"`
	PublishedAt string `json:"publishedAt"`
	Source      Source `json:"source"`
}

type Source struct {
	Name string `json:"name"`
}

type apiResponse struct {
	Articles []Article `json:"articles"`
}

func NewClient(apiKey string, baseUrl string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseUrl: baseUrl,
	}
}

func (c *Client) Search(query string) ([]Article, error) {
	endpoint := fmt.Sprintf("%s/everything?q=%s&apiKey=%s&pageSize=20",
		c.baseUrl, url.QueryEscape(query), c.apiKey)

	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return apiResp.Articles, nil
}
