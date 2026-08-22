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

type PresearchEngine struct {
	engine.BaseEngine
}

func NewPresearchEngine() *PresearchEngine {
	return &PresearchEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "presearch",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     0.9,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (p *PresearchEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
	}

	searchURL := fmt.Sprintf("https://engine.presearch.org/search?q=%s", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		searchURL += fmt.Sprintf("&page=%d", query.PageNo)
	}

	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	var items []types.SearchResult

	if err == nil && resp.StatusCode < 400 {
		doc, dErr := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if dErr == nil {
			doc.Find("style, script, svg").Remove()
			doc.Find("div.search-result, div[data-test='search-result'], div.result, article").Each(func(i int, s *goquery.Selection) {
				link := s.Find("a.result-title, h3 a, a").First()
				href, _ := link.Attr("href")
				title := strings.TrimSpace(link.Text())
				content := strings.TrimSpace(s.Find("p.result-description, div.description, p").First().Text())

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

	// Fallback to Brave/DDG if Presearch times out or rate limits
	if len(items) == 0 {
		braveURL := fmt.Sprintf("https://search.brave.com/search?q=%s&source=web", url.QueryEscape(query.Query))
		bResp, bBody, bErr := client.Get(ctx, braveURL, &network.RequestOptions{Headers: headers})
		if bErr == nil && bResp.StatusCode == 200 {
			bDoc, _ := goquery.NewDocumentFromReader(bytes.NewReader(bBody))
			if bDoc != nil {
				bDoc.Find("style, script, svg").Remove()
				bDoc.Find("div.snippet").Each(func(i int, s *goquery.Selection) {
					link := s.Find("a.heading-link, a").First()
					href, _ := link.Attr("href")
					title := strings.TrimSpace(s.Find("div.title, h2").First().Text())
					if title == "" {
						title = strings.TrimSpace(link.Text())
					}
					content := strings.TrimSpace(s.Find("div.snippet-description, div.content, p").First().Text())

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

	container.Extend(p.Name(), items)
	return nil
}
