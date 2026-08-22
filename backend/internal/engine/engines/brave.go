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

type BraveEngine struct {
	engine.BaseEngine
}

func NewBraveEngine() *BraveEngine {
	return &BraveEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "brave",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (b *BraveEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://search.brave.com/search?q=%s&source=web", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		searchURL += fmt.Sprintf("&offset=%d", query.PageNo-1)
	}

	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
	}

	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("brave returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	doc.Find("style, script, svg").Remove()

	var items []types.SearchResult
	doc.Find("div.snippet, div[data-type='web'], article").Each(func(i int, s *goquery.Selection) {
		link := s.Find("a.heading-link, a.result-header, h2 a, a").First()
		href, _ := link.Attr("href")
		title := strings.TrimSpace(s.Find(".title, h2, .heading").First().Text())
		if title == "" {
			title = strings.TrimSpace(link.Text())
		}

		content := strings.TrimSpace(s.Find(".snippet-description, .snippet-content, div.content, div[class*='description'], p").First().Text())
		if content == "" {
			content = strings.TrimSpace(s.Find("p").Text())
		}

		if href != "" && title != "" && strings.HasPrefix(href, "http") {
			items = append(items, types.SearchResult{
				URL:      href,
				Title:    title,
				Content:  content,
				Category: "general",
				Template: "default.html",
			})
		}
	})

	container.Extend(b.Name(), items)
	return nil
}
