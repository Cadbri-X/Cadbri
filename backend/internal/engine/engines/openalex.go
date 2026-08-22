package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"cadbri/internal/engine"
	"cadbri/internal/network"
	"cadbri/internal/results"
	"cadbri/internal/types"
)

type OpenAlexEngine struct {
	engine.BaseEngine
}

func NewOpenAlexEngine() *OpenAlexEngine {
	return &OpenAlexEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "openalex",
			EngineCategories: []string{"science", "academic"},
			EngineWeight:     1.0,
			EngineTimeout:    4 * time.Second,
			CanPage:          true,
		},
	}
}

type openAlexResponse struct {
	Meta struct {
		Count int64 `json:"count"`
	} `json:"meta"`
	Results []struct {
		ID             string `json:"id"`
		DOI            string `json:"doi"`
		Title          string `json:"title"`
		PublicationYear int   `json:"publication_year"`
		CitedByCount   int    `json:"cited_by_count"`
		Authorships    []struct {
			Author struct {
				DisplayName string `json:"display_name"`
			} `json:"author"`
		} `json:"authorships"`
	} `json:"results"`
}

func (o *OpenAlexEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	page := query.PageNo
	if page < 1 {
		page = 1
	}

	apiURL := fmt.Sprintf("https://api.openalex.org/works?search=%s&per-page=10&page=%d", url.QueryEscape(query.Query), page)
	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var data openAlexResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if data.Meta.Count > 0 {
		container.SetNumberOfResults(data.Meta.Count)
	}

	var items []types.SearchResult
	for _, it := range data.Results {
		var authorNames []string
		for _, auth := range it.Authorships {
			authorNames = append(authorNames, auth.Author.DisplayName)
		}
		authorsStr := strings.Join(authorNames, ", ")

		targetURL := it.DOI
		if targetURL == "" {
			targetURL = it.ID
		}

		desc := fmt.Sprintf("[%d | Citations: %d] %s", it.PublicationYear, it.CitedByCount, authorsStr)

		items = append(items, types.SearchResult{
			URL:      targetURL,
			Title:    it.Title,
			Content:  desc,
			Category: "science",
			Template: "paper.html",
			Author:   authorsStr,
		})
	}

	container.Extend(o.Name(), items)
	return nil
}
