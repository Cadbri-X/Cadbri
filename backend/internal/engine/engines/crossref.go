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

type CrossrefEngine struct {
	engine.BaseEngine
}

func NewCrossrefEngine() *CrossrefEngine {
	return &CrossrefEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "crossref",
			EngineCategories: []string{"science", "academic"},
			EngineWeight:     1.0,
			EngineTimeout:    4 * time.Second,
			CanPage:          true,
		},
	}
}

type crossrefResponse struct {
	Message struct {
		TotalResults int64 `json:"total-results"`
		Items        []struct {
			DOI    string   `json:"DOI"`
			Title  []string `json:"title"`
			URL    string   `json:"URL"`
			Author []struct {
				Given  string `json:"given"`
				Family string `json:"family"`
			} `json:"author"`
			ContainerTitle []string `json:"container-title"`
			Publisher      string   `json:"publisher"`
		} `json:"items"`
	} `json:"message"`
}

func (c *CrossrefEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	offset := 0
	if query.PageNo > 1 {
		offset = (query.PageNo - 1) * 10
	}

	apiURL := fmt.Sprintf(
		"https://api.crossref.org/works?query=%s&rows=10&offset=%d",
		url.QueryEscape(query.Query), offset,
	)

	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var data crossrefResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if data.Message.TotalResults > 0 {
		container.SetNumberOfResults(data.Message.TotalResults)
	}

	var items []types.SearchResult
	for _, it := range data.Message.Items {
		title := "Untitled"
		if len(it.Title) > 0 {
			title = it.Title[0]
		}

		var authors []string
		for _, auth := range it.Author {
			authors = append(authors, fmt.Sprintf("%s %s", auth.Given, auth.Family))
		}
		authStr := strings.Join(authors, ", ")

		journal := ""
		if len(it.ContainerTitle) > 0 {
			journal = it.ContainerTitle[0]
		}

		desc := fmt.Sprintf("[DOI: %s | %s] %s", it.DOI, journal, authStr)

		items = append(items, types.SearchResult{
			URL:      it.URL,
			Title:    title,
			Content:  desc,
			Category: "science",
			Template: "paper.html",
			Author:   authStr,
		})
	}

	container.Extend(c.Name(), items)
	return nil
}
