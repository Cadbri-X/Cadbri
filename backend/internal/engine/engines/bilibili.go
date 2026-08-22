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

type BilibiliEngine struct {
	engine.BaseEngine
}

func NewBilibiliEngine() *BilibiliEngine {
	return &BilibiliEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "bilibili",
			EngineCategories: []string{"videos", "media"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type bilibiliSearchResponse struct {
	Code int `json:"code"`
	Data struct {
		NumResults int64 `json:"numResults"`
		Result     []struct {
			Arcurl      string `json:"arcurl"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Pic         string `json:"pic"`
			Author      string `json:"author"`
			Play        int    `json:"play"`
		} `json:"result"`
	} `json:"data"`
}

func (b *BilibiliEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	page := query.PageNo
	if page < 1 {
		page = 1
	}

	apiURL := fmt.Sprintf(
		"https://api.bilibili.com/x/web-interface/search/type?search_type=video&keyword=%s&page=%d",
		url.QueryEscape(query.Query), page,
	)

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Referer":    "https://www.bilibili.com/",
	}

	_, body, err := client.Get(ctx, apiURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}

	var data bilibiliSearchResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if data.Data.NumResults > 0 {
		container.SetNumberOfResults(data.Data.NumResults)
	}

	var items []types.SearchResult
	for _, it := range data.Data.Result {
		cleanTitle := strings.ReplaceAll(it.Title, "<em class=\"keyword\">", "")
		cleanTitle = strings.ReplaceAll(cleanTitle, "</em>", "")

		picURL := it.Pic
		if strings.HasPrefix(picURL, "//") {
			picURL = "https:" + picURL
		}

		desc := fmt.Sprintf("[%s | Views: %d] %s", it.Author, it.Play, it.Description)

		items = append(items, types.SearchResult{
			URL:       it.Arcurl,
			Title:     cleanTitle,
			Content:   desc,
			Category:  "videos",
			Template:  "videos.html",
			Thumbnail: picURL,
			Author:    it.Author,
		})
	}

	container.Extend(b.Name(), items)
	return nil
}
