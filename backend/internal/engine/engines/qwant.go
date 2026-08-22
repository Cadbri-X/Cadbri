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

type QwantEngine struct {
	engine.BaseEngine
}

func NewQwantEngine() *QwantEngine {
	return &QwantEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "qwant",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type qwantResponse struct {
	Status string `json:"status"`
	Data   struct {
		Result struct {
			Total int64 `json:"total"`
			Items []struct {
				URL   string `json:"url"`
				Title string `json:"title"`
				Desc  string `json:"desc"`
			} `json:"items"`
		} `json:"result"`
	} `json:"data"`
}

func (q *QwantEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	offset := 0
	if query.PageNo > 1 {
		offset = (query.PageNo - 1) * 10
	}

	apiURL := fmt.Sprintf(
		"https://api.qwant.com/v3/search/web?q=%s&count=10&offset=%d&locale=en_US",
		url.QueryEscape(query.Query), offset,
	)

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Accept":     "application/json",
	}

	_, body, err := client.Get(ctx, apiURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}

	var data qwantResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if data.Data.Result.Total > 0 {
		container.SetNumberOfResults(data.Data.Result.Total)
	}

	var items []types.SearchResult
	for _, it := range data.Data.Result.Items {
		if it.URL != "" && it.Title != "" {
			items = append(items, types.SearchResult{
				URL:      it.URL,
				Title:    it.Title,
				Content:  it.Desc,
				Category: "general",
				Template: "default.html",
			})
		}
	}

	container.Extend(q.Name(), items)
	return nil
}
