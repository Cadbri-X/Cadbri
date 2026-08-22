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

type PinterestEngine struct {
	engine.BaseEngine
}

func NewPinterestEngine() *PinterestEngine {
	return &PinterestEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "pinterest",
			EngineCategories: []string{"images"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type pinterestResponse struct {
	ResourceData struct {
		Results []struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Images      struct {
				Orig struct {
					URL string `json:"url"`
				} `json:"orig"`
				Small struct {
					URL string `json:"url"`
				} `json:"236x"`
			} `json:"images"`
		} `json:"results"`
	} `json:"resource_response"`
}

func (p *PinterestEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	dataParam := fmt.Sprintf(`{"options":{"query":"%s","scope":"pins"},"context":{}}`, query.Query)
	apiURL := fmt.Sprintf(
		"https://www.pinterest.com/resource/BaseSearchResource/get/?source_url=/search/pins/?q=%s&data=%s",
		url.QueryEscape(query.Query), url.QueryEscape(dataParam),
	)

	headers := map[string]string{
		"X-Requested-With": "XMLHttpRequest",
		"Accept":           "application/json",
	}

	_, body, err := client.Get(ctx, apiURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}

	var data pinterestResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	var items []types.SearchResult
	for _, it := range data.ResourceData.Results {
		pinURL := fmt.Sprintf("https://www.pinterest.com/pin/%s/", it.ID)
		title := it.Title
		if title == "" {
			title = it.Description
		}
		if title == "" {
			title = "Pinterest Pin"
		}

		imgSrc := it.Images.Orig.URL
		thumb := it.Images.Small.URL
		if thumb == "" {
			thumb = imgSrc
		}

		if imgSrc != "" {
			items = append(items, types.SearchResult{
				URL:       pinURL,
				Title:     title,
				Content:   it.Description,
				Category:  "images",
				Template:  "images.html",
				ImgSrc:    imgSrc,
				Thumbnail: thumb,
			})
		}
	}

	container.Extend(p.Name(), items)
	return nil
}
