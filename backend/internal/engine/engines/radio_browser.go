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

type RadioBrowserEngine struct {
	engine.BaseEngine
}

func NewRadioBrowserEngine() *RadioBrowserEngine {
	return &RadioBrowserEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "radio_browser",
			EngineCategories: []string{"music", "media"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type radioStation struct {
	StationUUID string `json:"stationuuid"`
	Name        string `json:"name"`
	URL         string `json:"url_resolved"`
	Homepage    string `json:"homepage"`
	Favicon     string `json:"favicon"`
	Tags        string `json:"tags"`
	Country     string `json:"country"`
	Votes       int    `json:"votes"`
}

func (r *RadioBrowserEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	apiURL := fmt.Sprintf("https://de1.api.radio-browser.info/json/stations/byname/%s?limit=10", url.QueryEscape(query.Query))
	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var stations []radioStation
	if err := json.Unmarshal(body, &stations); err != nil {
		return err
	}

	var items []types.SearchResult
	for _, it := range stations {
		desc := fmt.Sprintf("[%s | Tags: %s | Votes: %d]", it.Country, it.Tags, it.Votes)
		targetURL := it.Homepage
		if targetURL == "" {
			targetURL = it.URL
		}

		items = append(items, types.SearchResult{
			URL:       targetURL,
			Title:     it.Name,
			Content:   desc,
			Category:  "music",
			Template:  "audio.html",
			Thumbnail: it.Favicon,
			AudioSrc:  it.URL,
		})
	}

	container.Extend(r.Name(), items)
	return nil
}
