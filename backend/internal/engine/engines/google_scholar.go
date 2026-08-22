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

type GoogleScholarEngine struct {
	engine.BaseEngine
}

func NewGoogleScholarEngine() *GoogleScholarEngine {
	return &GoogleScholarEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "google_scholar",
			EngineCategories: []string{"science", "academic"},
			EngineWeight:     1.1,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (g *GoogleScholarEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	start := 0
	if query.PageNo > 1 {
		start = (query.PageNo - 1) * 10
	}

	searchURL := fmt.Sprintf("https://scholar.google.com/scholar?q=%s&hl=en&start=%d", url.QueryEscape(query.Query), start)
	headers := map[string]string{
		"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	}

	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("google scholar returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("div.gs_r.gs_or.gs_scl").Each(func(i int, s *goquery.Selection) {
		link := s.Find("h3.gs_rt a").First()
		href, _ := link.Attr("href")
		title := strings.TrimSpace(link.Text())
		authors := strings.TrimSpace(s.Find("div.gs_a").First().Text())
		snippet := strings.TrimSpace(s.Find("div.gs_rs").First().Text())

		if href != "" && title != "" {
			content := snippet
			if authors != "" {
				content = fmt.Sprintf("[%s] %s", authors, snippet)
			}
			items = append(items, types.SearchResult{
				URL:      href,
				Title:    title,
				Content:  content,
				Category: "science",
				Template: "paper.html",
				Author:   authors,
			})
		}
	})

	container.Extend(g.Name(), items)
	return nil
}
