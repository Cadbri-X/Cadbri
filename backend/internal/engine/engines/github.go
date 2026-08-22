package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"cadbri/internal/engine"
	"cadbri/internal/network"
	"cadbri/internal/results"
	"cadbri/internal/types"
)

type GitHubEngine struct {
	engine.BaseEngine
}

func NewGitHubEngine() *GitHubEngine {
	return &GitHubEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "github",
			EngineCategories: []string{"it", "repos"},
			EngineWeight:     1.2,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type ghRepoResponse struct {
	TotalCount int64 `json:"total_count"`
	Items      []struct {
		FullName    string `json:"full_name"`
		HTMLURL     string `json:"html_url"`
		Description string `json:"description"`
		Stargazers  int64  `json:"stargazers_count"`
		Language    string `json:"language"`
		UpdatedAt   string `json:"updated_at"`
	} `json:"items"`
}

func (g *GitHubEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	page := query.PageNo
	if page < 1 {
		page = 1
	}

	apiURL := fmt.Sprintf("https://api.github.com/search/repositories?q=%s&sort=stars&order=desc&per_page=10&page=%d", url.QueryEscape(query.Query), page)
	headers := map[string]string{
		"Accept":     "application/vnd.github.v3+json",
		"User-Agent": "Cadbri-Go-Aggregator/1.0",
	}

	_, body, err := client.Get(ctx, apiURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}

	var data ghRepoResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if data.TotalCount > 0 {
		container.SetNumberOfResults(data.TotalCount)
	}

	var items []types.SearchResult
	for _, it := range data.Items {
		desc := it.Description
		if it.Language != "" {
			desc = fmt.Sprintf("[%s ⭐ %d] %s", it.Language, it.Stargazers, desc)
		}

		items = append(items, types.SearchResult{
			URL:      it.HTMLURL,
			Title:    it.FullName,
			Content:  desc,
			Category: "it",
			Template: "default.html",
			PubDate:  it.UpdatedAt,
		})
	}

	container.Extend(g.Name(), items)
	return nil
}
