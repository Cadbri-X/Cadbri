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

type GiteaEngine struct {
	engine.BaseEngine
}

func NewGiteaEngine() *GiteaEngine {
	return &GiteaEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "codeberg",
			EngineCategories: []string{"it", "repos"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type giteaSearchResponse struct {
	OK   bool `json:"ok"`
	Data []struct {
		FullName    string `json:"full_name"`
		HTMLURL     string `json:"html_url"`
		Description string `json:"description"`
		Stars       int64  `json:"stars_count"`
		Updated     string `json:"updated_at"`
	} `json:"data"`
}

func (g *GiteaEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	page := query.PageNo
	if page < 1 {
		page = 1
	}

	apiURL := fmt.Sprintf("https://codeberg.org/api/v1/repos/search?q=%s&limit=10&page=%d", url.QueryEscape(query.Query), page)
	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var data giteaSearchResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	var items []types.SearchResult
	for _, it := range data.Data {
		desc := it.Description
		if it.Stars > 0 {
			desc = fmt.Sprintf("[⭐ %d] %s", it.Stars, desc)
		}

		items = append(items, types.SearchResult{
			URL:      it.HTMLURL,
			Title:    it.FullName,
			Content:  desc,
			Category: "it",
			Template: "default.html",
			PubDate:  it.Updated,
		})
	}

	container.Extend(g.Name(), items)
	return nil
}
