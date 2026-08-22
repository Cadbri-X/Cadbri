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

type NPMEngine struct {
	engine.BaseEngine
}

func NewNPMEngine() *NPMEngine {
	return &NPMEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "npm",
			EngineCategories: []string{"it", "packages"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type npmSearchResponse struct {
	Total   int64 `json:"total"`
	Objects []struct {
		Package struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Description string `json:"description"`
			Date        string `json:"date"`
			Links       struct {
				NPM        string `json:"npm"`
				Repository string `json:"repository"`
			} `json:"links"`
			Publisher struct {
				Username string `json:"username"`
			} `json:"publisher"`
		} `json:"package"`
	} `json:"objects"`
}

func (n *NPMEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	from := 0
	if query.PageNo > 1 {
		from = (query.PageNo - 1) * 10
	}

	apiURL := fmt.Sprintf("https://registry.npmjs.org/-/v1/search?text=%s&size=10&from=%d", url.QueryEscape(query.Query), from)
	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var data npmSearchResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if data.Total > 0 {
		container.SetNumberOfResults(data.Total)
	}

	var items []types.SearchResult
	for _, obj := range data.Objects {
		pkg := obj.Package
		desc := fmt.Sprintf("[v%s by %s] %s", pkg.Version, pkg.Publisher.Username, pkg.Description)

		link := pkg.Links.NPM
		if link == "" {
			link = fmt.Sprintf("https://www.npmjs.com/package/%s", pkg.Name)
		}

		items = append(items, types.SearchResult{
			URL:      link,
			Title:    pkg.Name,
			Content:  desc,
			Category: "it",
			Template: "default.html",
			PubDate:  pkg.Date,
		})
	}

	container.Extend(n.Name(), items)
	return nil
}
