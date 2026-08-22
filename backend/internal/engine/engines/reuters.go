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

type ReutersEngine struct {
	engine.BaseEngine
}

func NewReutersEngine() *ReutersEngine {
	return &ReutersEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "reuters",
			EngineCategories: []string{"news"},
			EngineWeight:     1.1,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (r *ReutersEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://www.reuters.com/site-search/?query=%s", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		searchURL += fmt.Sprintf("&offset=%d", (query.PageNo-1)*10)
	}

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	}

	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("reuters returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("li[class*='search-results__item'], div.media-story-card").Each(func(i int, s *goquery.Selection) {
		link := s.Find("a[data-testid='Heading'], h3 a, a").First()
		href, _ := link.Attr("href")
		title := strings.TrimSpace(link.Text())
		pubTime := strings.TrimSpace(s.Find("time, span[class*='date']").First().Text())

		if href != "" && title != "" {
			if strings.HasPrefix(href, "/") {
				href = "https://www.reuters.com" + href
			}
			items = append(items, types.SearchResult{
				URL:      href,
				Title:    title,
				Content:  fmt.Sprintf("[Reuters News | %s]", pubTime),
				Category: "news",
				Template: "news.html",
				Author:   "Reuters",
				PubDate:  pubTime,
			})
		}
	})

	container.Extend(r.Name(), items)
	return nil
}
