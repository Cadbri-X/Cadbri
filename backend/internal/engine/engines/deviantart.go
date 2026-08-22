package engines

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"cadbri/internal/engine"
	"cadbri/internal/network"
	"cadbri/internal/results"
	"cadbri/internal/types"
)

type DeviantArtEngine struct {
	engine.BaseEngine
}

func NewDeviantArtEngine() *DeviantArtEngine {
	return &DeviantArtEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "deviantart",
			EngineCategories: []string{"images"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (d *DeviantArtEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://www.deviantart.com/search?q=%s", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		searchURL += fmt.Sprintf("&page=%d", query.PageNo)
	}

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	}

	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("deviantart returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("div[data-hook='deviation_std_thumb'], a[data-hook='deviation_link']").Each(func(i int, s *goquery.Selection) {
		link := s
		if !s.Is("a") {
			link = s.Find("a").First()
		}
		href, _ := link.Attr("href")
		img := s.Find("img").First()
		imgSrc, _ := img.Attr("src")
		alt, _ := img.Attr("alt")

		if href != "" && imgSrc != "" {
			title := alt
			if title == "" {
				title = "DeviantArt Artwork"
			}
			items = append(items, types.SearchResult{
				URL:       href,
				Title:     title,
				Content:   "Artwork on DeviantArt",
				Category:  "images",
				Template:  "images.html",
				ImgSrc:    imgSrc,
				Thumbnail: imgSrc,
			})
		}
	})

	container.Extend(d.Name(), items)
	return nil
}
