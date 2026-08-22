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

type StackExchangeEngine struct {
	engine.BaseEngine
}

func NewStackExchangeEngine() *StackExchangeEngine {
	return &StackExchangeEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "stackoverflow",
			EngineCategories: []string{"it", "q&a"},
			EngineWeight:     1.1,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type stackExchangeResponse struct {
	Items []struct {
		Title        string   `json:"title"`
		Link         string   `json:"link"`
		Score        int64    `json:"score"`
		AnswerCount  int64    `json:"answer_count"`
		IsAnswered   bool     `json:"is_answered"`
		Tags         []string `json:"tags"`
		CreationDate int64    `json:"creation_date"`
	} `json:"items"`
	Total int64 `json:"total"`
}

func (s *StackExchangeEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	page := query.PageNo
	if page < 1 {
		page = 1
	}

	apiURL := fmt.Sprintf(
		"https://api.stackexchange.com/2.3/search/advanced?q=%s&site=stackoverflow&pagesize=10&page=%d&filter=default",
		url.QueryEscape(query.Query), page,
	)

	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var data stackExchangeResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	var items []types.SearchResult
	for _, it := range data.Items {
		tagsStr := strings.Join(it.Tags, ", ")
		status := "Unanswered"
		if it.IsAnswered {
			status = "Answered"
		}
		desc := fmt.Sprintf("[%s | Votes: %d | Answers: %d] Tags: [%s]", status, it.Score, it.AnswerCount, tagsStr)

		items = append(items, types.SearchResult{
			URL:      it.Link,
			Title:    it.Title,
			Content:  desc,
			Category: "it",
			Template: "default.html",
		})
	}

	container.Extend(s.Name(), items)
	return nil
}
