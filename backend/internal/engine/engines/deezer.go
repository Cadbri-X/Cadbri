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

type DeezerEngine struct {
	engine.BaseEngine
}

func NewDeezerEngine() *DeezerEngine {
	return &DeezerEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "deezer",
			EngineCategories: []string{"music", "media"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type deezerResponse struct {
	Total int64 `json:"total"`
	Data  []struct {
		ID       int64  `json:"id"`
		Title    string `json:"title"`
		Link     string `json:"link"`
		Duration int    `json:"duration"`
		Preview  string `json:"preview"`
		Artist   struct {
			Name string `json:"name"`
		} `json:"artist"`
		Album struct {
			Title string `json:"title"`
			Cover string `json:"cover_medium"`
		} `json:"album"`
	} `json:"data"`
}

func (d *DeezerEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	index := 0
	if query.PageNo > 1 {
		index = (query.PageNo - 1) * 10
	}

	apiURL := fmt.Sprintf("https://api.deezer.com/search?q=%s&limit=10&index=%d", url.QueryEscape(query.Query), index)
	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var data deezerResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if data.Total > 0 {
		container.SetNumberOfResults(data.Total)
	}

	var items []types.SearchResult
	for _, it := range data.Data {
		desc := fmt.Sprintf("[%s - %s | Duration: %ds]", it.Artist.Name, it.Album.Title, it.Duration)

		items = append(items, types.SearchResult{
			URL:       it.Link,
			Title:     fmt.Sprintf("%s - %s", it.Artist.Name, it.Title),
			Content:   desc,
			Category:  "music",
			Template:  "audio.html",
			Thumbnail: it.Album.Cover,
			AudioSrc:  it.Preview,
			Length:    &it.Duration,
			Author:    it.Artist.Name,
		})
	}

	container.Extend(d.Name(), items)
	return nil
}
