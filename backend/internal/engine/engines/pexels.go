package engines

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"cadbri/internal/engine"
	"cadbri/internal/network"
	"cadbri/internal/results"
	"cadbri/internal/types"
)

type PexelsEngine struct {
	engine.BaseEngine
}

func NewPexelsEngine() *PexelsEngine {
	return &PexelsEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "pexels",
			EngineCategories: []string{"images"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (p *PexelsEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://www.pexels.com/search/%s/", url.PathEscape(query.Query))
	if query.PageNo > 1 {
		searchURL += fmt.Sprintf("?page=%d", query.PageNo)
	}

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	}

	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("pexels returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("article, div[data-testid='photo-card']").Each(func(i int, s *goquery.Selection) {
		link := s.Find("a").First()
		href, _ := link.Attr("href")
		img := s.Find("img").First()
		imgSrc, _ := img.Attr("src")
		alt, _ := img.Attr("alt")

		if href != "" && imgSrc != "" {
			if strings.HasPrefix(href, "/") {
				href = "https://www.pexels.com" + href
			}
			title := alt
			if title == "" {
				title = "Pexels Photo"
			}
			items = append(items, types.SearchResult{
				URL:       href,
				Title:     title,
				Content:   "Free stock photo on Pexels",
				Category:  "images",
				Template:  "images.html",
				ImgSrc:    imgSrc,
				Thumbnail: imgSrc,
			})
		}
	})

	container.Extend(p.Name(), items)
	return nil
}
