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

type TagesschauEngine struct {
	engine.BaseEngine
}

func NewTagesschauEngine() *TagesschauEngine {
	return &TagesschauEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "tagesschau",
			EngineCategories: []string{"news"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type tagesschauResponse struct {
	TotalCount int64 `json:"totalCount"`
	SearchResults []struct {
		Title       string `json:"title"`
		URL         string `json:"details"`
		SophoraID   string `json:"sophoraId"`
		Date        string `json:"date"`
		TeaserImage struct {
			Portraetgross struct {
				URL string `json:"url"`
			} `json:"portraetgross"`
		} `json:"teaserImage"`
	} `json:"searchResults"`
}

func (t *TagesschauEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	apiURL := fmt.Sprintf("https://www.tagesschau.de/api2u/search/?searchText=%s&pageSize=10", url.QueryEscape(query.Query))
	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var data tagesschauResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if data.TotalCount > 0 {
		container.SetNumberOfResults(data.TotalCount)
	}

	var items []types.SearchResult
	for _, it := range data.SearchResults {
		articleURL := it.URL
		if articleURL == "" && it.SophoraID != "" {
			articleURL = fmt.Sprintf("https://www.tagesschau.de/%s.html", it.SophoraID)
		}

		items = append(items, types.SearchResult{
			URL:       articleURL,
			Title:     it.Title,
			Content:   fmt.Sprintf("[Tagesschau | %s]", it.Date),
			Category:  "news",
			Template:  "news.html",
			Thumbnail: it.TeaserImage.Portraetgross.URL,
			Author:    "Tagesschau",
			PubDate:   it.Date,
		})
	}

	container.Extend(t.Name(), items)
	return nil
}
