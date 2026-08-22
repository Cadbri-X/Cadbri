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

type WikipediaEngine struct {
	engine.BaseEngine
}

func NewWikipediaEngine() *WikipediaEngine {
	return &WikipediaEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "wikipedia",
			EngineCategories: []string{"science"},
			EngineWeight:     1.5,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type wikiSearchResponse struct {
	Query struct {
		Searchinfo struct {
			Totalhits int64 `json:"totalhits"`
		} `json:"searchinfo"`
		Search []struct {
			Title     string `json:"title"`
			PageID    int64  `json:"pageid"`
			Snippet   string `json:"snippet"`
			Timestamp string `json:"timestamp"`
		} `json:"search"`
	} `json:"query"`
}

func (w *WikipediaEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	lang := "en"
	if query.Language != "" && query.Language != "all" {
		lang = strings.Split(query.Language, "-")[0]
	}

	offset := 0
	if query.PageNo > 1 {
		offset = (query.PageNo - 1) * 10
	}

	apiURL := fmt.Sprintf(
		"https://%s.wikipedia.org/w/api.php?action=query&list=search&srsearch=%s&sroffset=%d&srlimit=10&format=json",
		lang, url.QueryEscape(query.Query), offset,
	)

	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var data wikiSearchResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if data.Query.Searchinfo.Totalhits > 0 {
		container.SetNumberOfResults(data.Query.Searchinfo.Totalhits)
	}

	var items []types.SearchResult
	for i, item := range data.Query.Search {
		cleanSnippet := strings.ReplaceAll(item.Snippet, "<span class=\"searchmatch\">", "")
		cleanSnippet = strings.ReplaceAll(cleanSnippet, "</span>", "")

		pageURL := fmt.Sprintf("https://%s.wikipedia.org/wiki/%s", lang, url.PathEscape(strings.ReplaceAll(item.Title, " ", "_")))

		items = append(items, types.SearchResult{
			URL:      pageURL,
			Title:    item.Title + " - Wikipedia",
			Content:  cleanSnippet,
			Category: "general",
			Template: "default.html",
			PubDate:  item.Timestamp,
		})

		// If first result matches query closely, add as infobox/instant answer
		if i == 0 && query.PageNo <= 1 && len(cleanSnippet) > 20 {
			container.AddAnswer(types.Answer{
				Answer: cleanSnippet,
				URL:    pageURL,
				Engine: w.Name(),
			})
			container.AddInfobox(types.Infobox{
				Infobox: item.Title,
				ID:      pageURL,
				Content: cleanSnippet,
				Engine:  w.Name(),
				URLs: []types.InfoboxURL{
					{Title: "Wikipedia", URL: pageURL},
				},
			})
		}
	}

	container.Extend(w.Name(), items)
	return nil
}
