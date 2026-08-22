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

type ADSEngine struct {
	engine.BaseEngine
}

func NewADSEngine() *ADSEngine {
	return &ADSEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "ads",
			EngineCategories: []string{"science", "academic"},
			EngineWeight:     1.0,
			EngineTimeout:    4 * time.Second,
			CanPage:          true,
		},
	}
}

func (a *ADSEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://ui.adsabs.harvard.edu/search/q=%s", url.QueryEscape(query.Query))
	resp, body, err := client.Get(ctx, searchURL, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("ads returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("div.article-item, div.result-item").Each(func(i int, s *goquery.Selection) {
		link := s.Find("a.title-link, h3 a").First()
		href, _ := link.Attr("href")
		title := strings.TrimSpace(link.Text())
		authors := strings.TrimSpace(s.Find(".author-list, .authors").First().Text())
		abstract := strings.TrimSpace(s.Find(".abstract-text, .snippet").First().Text())

		if href != "" && title != "" {
			if strings.HasPrefix(href, "/") {
				href = "https://ui.adsabs.harvard.edu" + href
			}
			items = append(items, types.SearchResult{
				URL:      href,
				Title:    title,
				Content:  fmt.Sprintf("[%s] %s", authors, abstract),
				Category: "science",
				Template: "paper.html",
				Author:   authors,
			})
		}
	})

	container.Extend(a.Name(), items)
	return nil
}
