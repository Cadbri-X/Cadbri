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

type BlackleEngine struct {
	engine.BaseEngine
}

func NewBlackleEngine() *BlackleEngine {
	return &BlackleEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "blackle",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (b *BlackleEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	}

	searchURL := fmt.Sprintf("http://www.blackle.com/results/?q=%s", url.QueryEscape(query.Query))
	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	var items []types.SearchResult

	if err == nil && resp.StatusCode == 200 {
		doc, dErr := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if dErr == nil {
			doc.Find("div.g, div.result, div.gs-webResult").Each(func(i int, s *goquery.Selection) {
				link := s.Find("a.gs-title, a.l, a").First()
				href, _ := link.Attr("href")
				title := strings.TrimSpace(link.Text())
				content := strings.TrimSpace(s.Find("div.gs-snippet, div.s, p").First().Text())

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
		}
	}

	// Blackle uses Google Custom Search - fallback to Google / Startpage format
	if len(items) == 0 {
		spURL := fmt.Sprintf("https://www.startpage.com/sp/search?query=%s", url.QueryEscape(query.Query))
		sResp, sBody, sErr := client.Get(ctx, spURL, &network.RequestOptions{Headers: headers})
		if sErr == nil && sResp.StatusCode == 200 {
			sDoc, _ := goquery.NewDocumentFromReader(bytes.NewReader(sBody))
			if sDoc != nil {
				sDoc.Find("div.result, div.w-gl__result").Each(func(i int, s *goquery.Selection) {
					link := s.Find("a.result-title, a.w-gl__result-title, a.result-link").First()
					href, _ := link.Attr("href")
					title := strings.TrimSpace(link.Text())
					content := strings.TrimSpace(s.Find("p.result-snippet, p.w-gl__description").First().Text())

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
			}
		}
	}

	container.Extend(b.Name(), items)
	return nil
}
