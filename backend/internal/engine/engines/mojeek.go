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

type MojeekEngine struct {
	engine.BaseEngine
}

func NewMojeekEngine() *MojeekEngine {
	return &MojeekEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "mojeek",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (m *MojeekEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://www.mojeek.com/search?q=%s", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		s := (query.PageNo-1)*10 + 1
		searchURL += fmt.Sprintf("&s=%d", s)
	}

	resp, body, err := client.Get(ctx, searchURL, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("mojeek returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("ul.results-standard li, ul.results li").Each(func(i int, s *goquery.Selection) {
		link := s.Find("a.title, a.ob").First()
		href, _ := link.Attr("href")
		title := strings.TrimSpace(link.Text())
		content := strings.TrimSpace(s.Find("p.s, p.snippet, div.s").First().Text())

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

	container.Extend(m.Name(), items)
	return nil
}
