package engines

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
	"time"

	"cadbri/internal/engine"
	"cadbri/internal/network"
	"cadbri/internal/results"
	"cadbri/internal/types"
)

type ArXivEngine struct {
	engine.BaseEngine
}

func NewArXivEngine() *ArXivEngine {
	return &ArXivEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "arxiv",
			EngineCategories: []string{"science", "academic"},
			EngineWeight:     1.1,
			EngineTimeout:    4 * time.Second,
			CanPage:          true,
		},
	}
}

type arxivFeed struct {
	XMLName xml.Name     `xml:"feed"`
	Entries []arxivEntry `xml:"entry"`
}

type arxivEntry struct {
	ID        string `xml:"id"`
	Published string `xml:"published"`
	Title     string `xml:"title"`
	Summary   string `xml:"summary"`
	Authors   []struct {
		Name string `xml:"name"`
	} `xml:"author"`
}

func (a *ArXivEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	start := 0
	if query.PageNo > 1 {
		start = (query.PageNo - 1) * 10
	}

	apiURL := fmt.Sprintf(
		"https://export.arxiv.org/api/query?search_query=all:%s&start=%d&max_results=10&sortBy=relevance",
		url.QueryEscape(query.Query), start,
	)

	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var feed arxivFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return err
	}

	var items []types.SearchResult
	for _, entry := range feed.Entries {
		var authorNames []string
		for _, auth := range entry.Authors {
			authorNames = append(authorNames, auth.Name)
		}
		authorsStr := strings.Join(authorNames, ", ")

		cleanTitle := strings.TrimSpace(strings.ReplaceAll(entry.Title, "\n", " "))
		cleanSummary := strings.TrimSpace(strings.ReplaceAll(entry.Summary, "\n", " "))

		desc := cleanSummary
		if authorsStr != "" {
			desc = fmt.Sprintf("[%s] %s", authorsStr, cleanSummary)
		}

		items = append(items, types.SearchResult{
			URL:      entry.ID,
			Title:    cleanTitle,
			Content:  desc,
			Category: "science",
			Template: "paper.html",
			Author:   authorsStr,
			PubDate:  entry.Published,
		})
	}

	container.Extend(a.Name(), items)
	return nil
}
