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

type UnsplashEngine struct {
	engine.BaseEngine
}

func NewUnsplashEngine() *UnsplashEngine {
	return &UnsplashEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "unsplash",
			EngineCategories: []string{"images"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type unsplashResponse struct {
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	Results    []struct {
		ID             string `json:"id"`
		Description    string `json:"description"`
		AltDescription string `json:"alt_description"`
		Links          struct {
			HTML string `json:"html"`
		} `json:"links"`
		Urls struct {
			Regular string `json:"regular"`
			Small   string `json:"small"`
			Thumb   string `json:"thumb"`
		} `json:"urls"`
		User struct {
			Name string `json:"name"`
		} `json:"user"`
	} `json:"results"`
}

func (u *UnsplashEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	page := query.PageNo
	if page < 1 {
		page = 1
	}

	apiURL := fmt.Sprintf("https://unsplash.com/napi/search/photos?query=%s&per_page=10&page=%d", url.QueryEscape(query.Query), page)
	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var data unsplashResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if data.Total > 0 {
		container.SetNumberOfResults(data.Total)
	}

	var items []types.SearchResult
	for _, it := range data.Results {
		title := it.AltDescription
		if title == "" {
			title = it.Description
		}
		if title == "" {
			title = fmt.Sprintf("Photo by %s", it.User.Name)
		}

		desc := fmt.Sprintf("Photo by %s on Unsplash", it.User.Name)

		items = append(items, types.SearchResult{
			URL:       it.Links.HTML,
			Title:     title,
			Content:   desc,
			Category:  "images",
			Template:  "images.html",
			ImgSrc:    it.Urls.Regular,
			Thumbnail: it.Urls.Small,
			Author:    it.User.Name,
		})
	}

	container.Extend(u.Name(), items)
	return nil
}
