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

type PkgGoDevEngine struct {
	engine.BaseEngine
}

func NewPkgGoDevEngine() *PkgGoDevEngine {
	return &PkgGoDevEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "pkg_go_dev",
			EngineCategories: []string{"it", "packages"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (p *PkgGoDevEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://pkg.go.dev/search?q=%s", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		searchURL += fmt.Sprintf("&page=%d", query.PageNo)
	}

	resp, body, err := client.Get(ctx, searchURL, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("pkg.go.dev returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("div.SearchSnippet").Each(func(i int, s *goquery.Selection) {
		link := s.Find("h2 a, a.SearchSnippet-header").First()
		href, _ := link.Attr("href")
		title := strings.TrimSpace(link.Text())
		content := strings.TrimSpace(s.Find("p.SearchSnippet-synopsis").First().Text())

		if href != "" && title != "" {
			fullURL := "https://pkg.go.dev" + href
			items = append(items, types.SearchResult{
				URL:      fullURL,
				Title:    title,
				Content:  content,
				Category: "it",
				Template: "default.html",
			})
		}
	})

	container.Extend(p.Name(), items)
	return nil
}
