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

type PyPIEngine struct {
	engine.BaseEngine
}

func NewPyPIEngine() *PyPIEngine {
	return &PyPIEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "pypi",
			EngineCategories: []string{"it", "packages"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (p *PyPIEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://pypi.org/search/?q=%s", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		searchURL += fmt.Sprintf("&page=%d", query.PageNo)
	}

	resp, body, err := client.Get(ctx, searchURL, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("pypi returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("a.package-snippet").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		name := strings.TrimSpace(s.Find(".package-snippet__name").Text())
		version := strings.TrimSpace(s.Find(".package-snippet__version").Text())
		desc := strings.TrimSpace(s.Find(".package-snippet__description").Text())
		pubDate, _ := s.Find("time").Attr("datetime")

		if href != "" && name != "" {
			fullURL := "https://pypi.org" + href
			title := name
			if version != "" {
				title = fmt.Sprintf("%s (%s)", name, version)
			}

			items = append(items, types.SearchResult{
				URL:      fullURL,
				Title:    title,
				Content:  desc,
				Category: "it",
				Template: "default.html",
				PubDate:  pubDate,
			})
		}
	})

	container.Extend(p.Name(), items)
	return nil
}
