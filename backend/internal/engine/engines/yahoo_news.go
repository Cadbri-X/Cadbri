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

type YahooNewsEngine struct {
	engine.BaseEngine
}

func NewYahooNewsEngine() *YahooNewsEngine {
	return &YahooNewsEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "yahoo_news",
			EngineCategories: []string{"news"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (y *YahooNewsEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://news.search.yahoo.com/search?p=%s", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		b := (query.PageNo-1)*10 + 1
		searchURL += fmt.Sprintf("&b=%d", b)
	}

	resp, body, err := client.Get(ctx, searchURL, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("yahoo news returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("ol.searchCenterMiddle li, div.NewsArticle").Each(func(i int, s *goquery.Selection) {
		link := s.Find("h4.s-title a, a.thmb").First()
		href, _ := link.Attr("href")
		title := strings.TrimSpace(link.Text())
		snippet := strings.TrimSpace(s.Find("p.s-desc, div.compText").First().Text())
		source := strings.TrimSpace(s.Find("span.s-source, cite").First().Text())
		pubTime := strings.TrimSpace(s.Find("span.s-time").First().Text())

		if href != "" && title != "" {
			if strings.Contains(href, "/RU=") {
				parts := strings.Split(href, "/RU=")
				if len(parts) > 1 {
					sub := strings.Split(parts[1], "/RK=")[0]
					if decoded, err := url.QueryUnescape(sub); err == nil {
						href = decoded
					}
				}
			}

			desc := snippet
			if source != "" {
				desc = fmt.Sprintf("[%s | %s] %s", source, pubTime, snippet)
			}

			items = append(items, types.SearchResult{
				URL:      href,
				Title:    title,
				Content:  desc,
				Category: "news",
				Template: "news.html",
				Author:   source,
				PubDate:  pubTime,
			})
		}
	})

	container.Extend(y.Name(), items)
	return nil
}
