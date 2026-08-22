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

type SogouEngine struct {
	engine.BaseEngine
}

func NewSogouEngine() *SogouEngine {
	return &SogouEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "sogou",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     0.8,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (s *SogouEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://www.sogou.com/web?query=%s", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		searchURL += fmt.Sprintf("&page=%d", query.PageNo)
	}

	headers := map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
	}

	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("sogou returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("div.vrwrap, div.rb, div.results > div").Each(func(i int, sel *goquery.Selection) {
		link := sel.Find("h3 a, a.title").First()
		href, _ := link.Attr("href")
		title := strings.TrimSpace(link.Text())
		content := strings.TrimSpace(sel.Find("div.txt-box, div.str-box, p.star-wiki").First().Text())

		if href != "" && title != "" {
			if strings.HasPrefix(href, "/link?url=") {
				href = "https://www.sogou.com" + href
			}
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
