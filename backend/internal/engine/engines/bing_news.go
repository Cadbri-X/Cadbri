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

type BingNewsEngine struct {
	engine.BaseEngine
}

func NewBingNewsEngine() *BingNewsEngine {
	return &BingNewsEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "bing_news",
			EngineCategories: []string{"news"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (b *BingNewsEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://www.bing.com/news/search?q=%s", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		first := (query.PageNo-1)*10 + 1
		searchURL += fmt.Sprintf("&first=%d", first)
	}

	headers := map[string]string{
		"Accept-Language": "en-US,en;q=0.9",
	}

	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("bing news returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("div.news-card, div.ans_nws").Each(func(i int, s *goquery.Selection) {
		link := s.Find("a.title, a.href").First()
		href, _ := link.Attr("href")
		title := strings.TrimSpace(link.Text())
		snippet := strings.TrimSpace(s.Find("div.snippet, p").First().Text())
		source := strings.TrimSpace(s.Find("div.source, span.source").First().Text())
		pubDate := strings.TrimSpace(s.Find("span[tabindex='0'], span.pubtime").First().Text())
		imgSrc, _ := s.Find("img").Attr("src")

		if href != "" && title != "" {
			desc := snippet
			if source != "" {
				desc = fmt.Sprintf("[%s | %s] %s", source, pubDate, snippet)
			}

			items = append(items, types.SearchResult{
				URL:       href,
				Title:     title,
				Content:   desc,
				Category:  "news",
				Template:  "news.html",
				Thumbnail: imgSrc,
				Author:    source,
				PubDate:   pubDate,
			})
		}
	})

	container.Extend(b.Name(), items)
	return nil
}
