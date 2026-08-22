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

type WibyEngine struct {
	engine.BaseEngine
}

func NewWibyEngine() *WibyEngine {
	return &WibyEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "wiby",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     0.8,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type wibyItem struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

func (w *WibyEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://wiby.me/json/?q=%s", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		searchURL += fmt.Sprintf("&p=%d", query.PageNo)
	}

	_, body, err := client.Get(ctx, searchURL, nil)
	if err != nil {
		return err
	}

	var rawList []wibyItem
	if err := json.Unmarshal(body, &rawList); err != nil {
		return err
	}

	var items []types.SearchResult
	for _, it := range rawList {
		if it.URL != "" && it.Title != "" {
			items = append(items, types.SearchResult{
				URL:      it.URL,
				Title:    it.Title,
				Content:  it.Snippet,
				Category: "general",
				Template: "default.html",
			})
		}
	}

	container.Extend(w.Name(), items)
	return nil
}
