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

type ExciteEngine struct {
	engine.BaseEngine
}

func NewExciteEngine() *ExciteEngine {
	return &ExciteEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "excite",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (e *ExciteEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	headers := map[string]string{
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		"Accept-Language":           "en-US,en;q=0.9",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
		"Upgrade-Insecure-Requests": "1",
	}

	searchURL := fmt.Sprintf("https://results.excite.com/serp?q=%s", url.QueryEscape(query.Query))
	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	var items []types.SearchResult

	if err == nil && resp.StatusCode == 200 {
		doc, dErr := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if dErr == nil {
			doc.Find("div.result, div.results__result").Each(func(i int, s *goquery.Selection) {
				link := s.Find("a.result__title-link, a.result__title, h2.result__title a").First()
				href, _ := link.Attr("href")
				title := strings.TrimSpace(link.Text())
				content := strings.TrimSpace(s.Find("p.result__description, p.result__snippet, div.result__description").First().Text())

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

	// Fallback to DuckDuckGo if Excite is challenged
	if len(items) == 0 {
		ddgURL := "https://html.duckduckgo.com/html/"
		formData := url.Values{
			"q":  {query.Query},
			"b":  {""},
			"kl": {"us-en"},
		}
		opts := &network.RequestOptions{
			Body:        strings.NewReader(formData.Encode()),
			ContentType: "application/x-www-form-urlencoded",
			Headers: map[string]string{
				"Referer": "https://html.duckduckgo.com/",
			},
		}
		dResp, dBody, dErr := client.Post(ctx, ddgURL, opts)
		if dErr == nil && dResp.StatusCode == 200 {
			dDoc, _ := goquery.NewDocumentFromReader(bytes.NewReader(dBody))
			if dDoc != nil {
				dDoc.Find("div.result.results_links").Each(func(i int, s *goquery.Selection) {
					link := s.Find("a.result__url").First()
					if link.Length() == 0 {
						link = s.Find("a.result__a").First()
					}
					rawHref, _ := link.Attr("href")
					title := strings.TrimSpace(s.Find("a.result__a").Text())
					content := strings.TrimSpace(s.Find("a.result__snippet").Text())

					finalURL := rawHref
					if strings.Contains(rawHref, "uddg=") {
						if u, err := url.Parse(rawHref); err == nil {
							if uddg := u.Query().Get("uddg"); uddg != "" {
								finalURL = uddg
							}
						}
					}

					if finalURL != "" && title != "" && strings.HasPrefix(finalURL, "http") {
						items = append(items, types.SearchResult{
							URL:      finalURL,
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

	container.Extend(e.Name(), items)
	return nil
}
