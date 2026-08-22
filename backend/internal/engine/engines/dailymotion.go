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

type DailymotionEngine struct {
	engine.BaseEngine
}

func NewDailymotionEngine() *DailymotionEngine {
	return &DailymotionEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "dailymotion",
			EngineCategories: []string{"videos", "media"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type dmResponse struct {
	Page    int  `json:"page"`
	Total   int64 `json:"total"`
	HasMore bool `json:"has_more"`
	List    []struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
		Thumbnail   string `json:"thumbnail_240_url"`
		Duration    int    `json:"duration"`
	} `json:"list"`
}

func (d *DailymotionEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	page := query.PageNo
	if page < 1 {
		page = 1
	}

	apiURL := fmt.Sprintf(
		"https://api.dailymotion.com/videos?search=%s&limit=10&page=%d&fields=id,title,description,url,thumbnail_240_url,duration",
		url.QueryEscape(query.Query), page,
	)

	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var data dmResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if data.Total > 0 {
		container.SetNumberOfResults(data.Total)
	}

	var items []types.SearchResult
	for _, it := range data.List {
		embedURL := fmt.Sprintf("https://www.dailymotion.com/embed/video/%s", it.ID)
		items = append(items, types.SearchResult{
			URL:       it.URL,
			Title:     it.Title,
			Content:   it.Description,
			Category:  "videos",
			Template:  "videos.html",
			Thumbnail: it.Thumbnail,
			IframeSrc: embedURL,
			Length:    &it.Duration,
		})
	}

	container.Extend(d.Name(), items)
	return nil
}
