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

type SeznamEngine struct {
	engine.BaseEngine
}

func NewSeznamEngine() *SeznamEngine {
	return &SeznamEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "seznam",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     0.9,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (s *SeznamEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://search.seznam.cz/?q=%s", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		from := (query.PageNo-1)*10 + 1
		searchURL += fmt.Sprintf("&from=%d", from)
	}

	resp, body, err := client.Get(ctx, searchURL, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("seznam returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("div.Result, div[data-dot='result']").Each(func(i int, sel *goquery.Selection) {
		link := sel.Find("a.Result-title, a.Result-titleLink").First()
		href, _ := link.Attr("href")
		title := strings.TrimSpace(link.Text())
		content := strings.TrimSpace(sel.Find("p.Result-text, div.Result-text").First().Text())

		if href != "" && title != "" {
			items = append(items, types.SearchResult{
				URL:      href,
				Title:    title,
				Content:  content,
				Category: "general",
				Template: "default.html",
			})
		}
	})

	container.Extend(s.Name(), items)
	return nil
}
