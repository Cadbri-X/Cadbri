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

type RubyGemsEngine struct {
	engine.BaseEngine
}

func NewRubyGemsEngine() *RubyGemsEngine {
	return &RubyGemsEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "rubygems",
			EngineCategories: []string{"it", "packages"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type rubyGemItem struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Info      string `json:"info"`
	ProjectURI string `json:"project_uri"`
	Downloads int64  `json:"downloads"`
}

func (r *RubyGemsEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	page := query.PageNo
	if page < 1 {
		page = 1
	}

	apiURL := fmt.Sprintf("https://rubygems.org/api/v1/search.json?query=%s&page=%d", url.QueryEscape(query.Query), page)
	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var gems []rubyGemItem
	if err := json.Unmarshal(body, &gems); err != nil {
		return err
	}

	var items []types.SearchResult
	for _, g := range gems {
		desc := fmt.Sprintf("[v%s 📥 %d] %s", g.Version, g.Downloads, g.Info)
		items = append(items, types.SearchResult{
			URL:      g.ProjectURI,
			Title:    g.Name,
			Content:  desc,
			Category: "it",
			Template: "default.html",
		})
	}

	container.Extend(r.Name(), items)
	return nil
}
