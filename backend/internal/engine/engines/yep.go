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

type YepEngine struct {
	engine.BaseEngine
}

func NewYepEngine() *YepEngine {
	return &YepEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "yep",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (y *YepEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
	}

	searchURL := fmt.Sprintf("https://yep.com/web?q=%s", url.QueryEscape(query.Query))
	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	var items []types.SearchResult

	if err == nil && resp.StatusCode == 200 {
		doc, dErr := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if dErr == nil {
			doc.Find("div.snippet, article, div.result").Each(func(i int, s *goquery.Selection) {
				link := s.Find("h2 a, a.title, a").First()
				href, _ := link.Attr("href")
				title := strings.TrimSpace(link.Text())
				content := strings.TrimSpace(s.Find("p.snippet, p.description, p").First().Text())

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

	// Fallback to Brave if Yep bot protection triggered
	if len(items) == 0 {
		braveURL := fmt.Sprintf("https://search.brave.com/search?q=%s&source=web", url.QueryEscape(query.Query))
		bResp, bBody, bErr := client.Get(ctx, braveURL, &network.RequestOptions{Headers: headers})
		if bErr == nil && bResp.StatusCode == 200 {
			bDoc, _ := goquery.NewDocumentFromReader(bytes.NewReader(bBody))
			if bDoc != nil {
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

	container.Extend(y.Name(), items)
	return nil
}
