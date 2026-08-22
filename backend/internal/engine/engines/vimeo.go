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

type VimeoEngine struct {
	engine.BaseEngine
}

func NewVimeoEngine() *VimeoEngine {
	return &VimeoEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "vimeo",
			EngineCategories: []string{"videos", "media"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (v *VimeoEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://vimeo.com/search?q=%s", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		searchURL += fmt.Sprintf("&page=%d", query.PageNo)
	}

	resp, body, err := client.Get(ctx, searchURL, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("vimeo returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("div.iris_video-vital, a.iris_link--gray-3, div[data-testid='video-card']").Each(func(i int, s *goquery.Selection) {
		link := s.Find("a").First()
		href, _ := link.Attr("href")
		title := strings.TrimSpace(s.Find(".iris_video-vital__title, h4").First().Text())
		content := strings.TrimSpace(s.Find(".iris_video-vital__description, p").First().Text())
		thumb, _ := s.Find("img").Attr("src")

		if href != "" && title != "" {
			if strings.HasPrefix(href, "/") {
				href = "https://vimeo.com" + href
			}
			items = append(items, types.SearchResult{
				URL:       href,
				Title:     title,
				Content:   content,
				Category:  "videos",
				Template:  "videos.html",
				Thumbnail: thumb,
			})
		}
	})

	container.Extend(v.Name(), items)
	return nil
}
