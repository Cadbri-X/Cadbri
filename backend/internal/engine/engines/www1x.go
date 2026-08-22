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

type Www1xEngine struct {
	engine.BaseEngine
}

func NewWww1xEngine() *Www1xEngine {
	return &Www1xEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "1x",
			EngineCategories: []string{"images"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (w *Www1xEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://1x.com/search/%s", url.PathEscape(query.Query))
	resp, body, err := client.Get(ctx, searchURL, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("1x returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("div.photo-item, a.photo-link, div.thumbnail").Each(func(i int, s *goquery.Selection) {
		link := s
		if !s.Is("a") {
			link = s.Find("a").First()
		}
		href, _ := link.Attr("href")
		img := s.Find("img").First()
		imgSrc, _ := img.Attr("src")
		alt, _ := img.Attr("alt")

		if href != "" && imgSrc != "" {
			if strings.HasPrefix(href, "/") {
				href = "https://1x.com" + href
			}
			title := alt
			if title == "" {
				title = "1x Photography"
			}
			items = append(items, types.SearchResult{
				URL:       href,
				Title:     title,
				Content:   "Curated photography on 1x.com",
				Category:  "images",
				Template:  "images.html",
				ImgSrc:    imgSrc,
				Thumbnail: imgSrc,
			})
		}
	})

	container.Extend(w.Name(), items)
	return nil
}
