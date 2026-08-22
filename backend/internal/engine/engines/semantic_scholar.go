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

type SemanticScholarEngine struct {
	engine.BaseEngine
}

func NewSemanticScholarEngine() *SemanticScholarEngine {
	return &SemanticScholarEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "semantic_scholar",
			EngineCategories: []string{"science", "academic"},
			EngineWeight:     1.1,
			EngineTimeout:    4 * time.Second,
			CanPage:          true,
		},
	}
}

type semanticScholarResponse struct {
	Total int64 `json:"total"`
	Data  []struct {
		PaperID       string `json:"paperId"`
		Title         string `json:"title"`
		URL           string `json:"url"`
		Abstract      string `json:"abstract"`
		Year          int    `json:"year"`
		CitationCount int    `json:"citationCount"`
		Authors       []struct {
			Name string `json:"name"`
		} `json:"authors"`
	} `json:"data"`
}

func (s *SemanticScholarEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	offset := 0
	if query.PageNo > 1 {
		offset = (query.PageNo - 1) * 10
	}

	apiURL := fmt.Sprintf(
		"https://api.semanticscholar.org/graph/v1/paper/search?query=%s&limit=10&offset=%d&fields=title,url,abstract,authors,year,citationCount",
		url.QueryEscape(query.Query), offset,
	)

	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var data semanticScholarResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if data.Total > 0 {
		container.SetNumberOfResults(data.Total)
	}

	var items []types.SearchResult
	for _, it := range data.Data {
		var authorNames []string
		for _, auth := range it.Authors {
			authorNames = append(authorNames, auth.Name)
		}
		authorsStr := strings.Join(authorNames, ", ")

		desc := it.Abstract
		if authorsStr != "" || it.Year > 0 {
			desc = fmt.Sprintf("[%d | Citations: %d | %s] %s", it.Year, it.CitationCount, authorsStr, it.Abstract)
		}

		paperURL := it.URL
		if paperURL == "" {
			paperURL = fmt.Sprintf("https://www.semanticscholar.org/paper/%s", it.PaperID)
		}

		items = append(items, types.SearchResult{
			URL:      paperURL,
			Title:    it.Title,
			Content:  desc,
			Category: "science",
			Template: "paper.html",
			Author:   authorsStr,
		})
	}

	container.Extend(s.Name(), items)
	return nil
}
