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

type CratesEngine struct {
	engine.BaseEngine
}

func NewCratesEngine() *CratesEngine {
	return &CratesEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "crates",
			EngineCategories: []string{"it", "packages"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type cratesResponse struct {
	Meta struct {
		Total int64 `json:"total"`
	} `json:"meta"`
	Crates []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		MaxVersion  string `json:"max_version"`
		Description string `json:"description"`
		Downloads   int64  `json:"downloads"`
		UpdatedAt   string `json:"updated_at"`
		Homepage    string `json:"homepage"`
	} `json:"crates"`
}

func (c *CratesEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	page := query.PageNo
	if page < 1 {
		page = 1
	}

	apiURL := fmt.Sprintf("https://crates.io/api/v1/crates?q=%s&per_page=10&page=%d", url.QueryEscape(query.Query), page)
	headers := map[string]string{
		"User-Agent": "Cadbri-Go-Aggregator/1.0",
	}

	_, body, err := client.Get(ctx, apiURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}

	var data cratesResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if data.Meta.Total > 0 {
		container.SetNumberOfResults(data.Meta.Total)
	}

	var items []types.SearchResult
	for _, cr := range data.Crates {
		desc := fmt.Sprintf("[v%s 📥 %d] %s", cr.MaxVersion, cr.Downloads, cr.Description)
		crateURL := fmt.Sprintf("https://crates.io/crates/%s", cr.ID)

		items = append(items, types.SearchResult{
			URL:      crateURL,
			Title:    cr.Name,
			Content:  desc,
			Category: "it",
			Template: "default.html",
			PubDate:  cr.UpdatedAt,
		})
	}

	container.Extend(c.Name(), items)
	return nil
}
