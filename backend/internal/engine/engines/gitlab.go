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

type GitLabEngine struct {
	engine.BaseEngine
}

func NewGitLabEngine() *GitLabEngine {
	return &GitLabEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "gitlab",
			EngineCategories: []string{"it", "repos"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type gitlabProject struct {
	NameWithNamespace string `json:"name_with_namespace"`
	WebURL            string `json:"web_url"`
	Description       string `json:"description"`
	StarCount         int64  `json:"star_count"`
	LastActivityAt    string `json:"last_activity_at"`
}

func (g *GitLabEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	page := query.PageNo
	if page < 1 {
		page = 1
	}

	apiURL := fmt.Sprintf("https://gitlab.com/api/v4/projects?search=%s&order_by=similarity&per_page=10&page=%d", url.QueryEscape(query.Query), page)
	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var projects []gitlabProject
	if err := json.Unmarshal(body, &projects); err != nil {
		return err
	}

	var items []types.SearchResult
	for _, it := range projects {
		desc := it.Description
		if it.StarCount > 0 {
			desc = fmt.Sprintf("[⭐ %d] %s", it.StarCount, desc)
		}

		items = append(items, types.SearchResult{
			URL:      it.WebURL,
			Title:    it.NameWithNamespace,
			Content:  desc,
			Category: "it",
			Template: "default.html",
			PubDate:  it.LastActivityAt,
		})
	}

	container.Extend(g.Name(), items)
	return nil
}
