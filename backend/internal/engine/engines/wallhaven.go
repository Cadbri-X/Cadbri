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

type WallhavenEngine struct {
	engine.BaseEngine
}

func NewWallhavenEngine() *WallhavenEngine {
	return &WallhavenEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "wallhaven",
			EngineCategories: []string{"images"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type wallhavenResponse struct {
	Data []struct {
		ID         string   `json:"id"`
		URL        string   `json:"url"`
		Path       string   `json:"path"`
		Resolution string   `json:"resolution"`
		Category   string   `json:"category"`
		Views      int      `json:"views"`
		Favorites  int      `json:"favorites"`
		Thumbs     struct {
			Large    string `json:"large"`
			Original string `json:"original"`
			Small    string `json:"small"`
		} `json:"thumbs"`
	} `json:"data"`
	Meta struct {
		Total int64 `json:"total"`
	} `json:"meta"`
}

func (w *WallhavenEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	page := query.PageNo
	if page < 1 {
		page = 1
	}

	apiURL := fmt.Sprintf("https://wallhaven.cc/api/v1/search?q=%s&page=%d", url.QueryEscape(query.Query), page)
	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var data wallhavenResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if data.Meta.Total > 0 {
		container.SetNumberOfResults(data.Meta.Total)
	}

	var items []types.SearchResult
	for _, it := range data.Data {
		desc := fmt.Sprintf("[%s | %s | 👁️ %d | ❤️ %d]", it.Resolution, it.Category, it.Views, it.Favorites)
		thumb := it.Thumbs.Large
		if thumb == "" {
			thumb = it.Thumbs.Small
		}

		items = append(items, types.SearchResult{
			URL:       it.URL,
			Title:     fmt.Sprintf("Wallpaper %s (%s)", it.ID, it.Resolution),
			Content:   desc,
			Category:  "images",
			Template:  "images.html",
			ImgSrc:    it.Path,
			Thumbnail: thumb,
		})
	}

	container.Extend(w.Name(), items)
	return nil
}
