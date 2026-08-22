package engines

import (
	"bytes"
	"context"
	"encoding/base64"
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

type WebCrawlerEngine struct {
	engine.BaseEngine
}

func NewWebCrawlerEngine() *WebCrawlerEngine {
	return &WebCrawlerEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "webcrawler",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (w *WebCrawlerEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
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

	searchURL := fmt.Sprintf("https://www.webcrawler.com/serp?q=%s", url.QueryEscape(query.Query))
	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	var items []types.SearchResult

	if err == nil && resp.StatusCode == 200 {
		doc, dErr := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if dErr == nil {
			doc.Find("style, script, svg").Remove()
			doc.Find("div.result, div.results__result, div.result-item").Each(func(i int, s *goquery.Selection) {
				link := s.Find("a.result__title-link, a.result__title, h2.result__title a, a").First()
				href, _ := link.Attr("href")
				title := strings.TrimSpace(link.Text())
				content := strings.TrimSpace(s.Find("p.result__description, p.result__snippet, div.result__description, div.result__body, p").First().Text())

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

	// Fallback to Bing if WebCrawler is Cloudflare-rate limited
	if len(items) == 0 {
		bingURL := fmt.Sprintf("https://www.bing.com/search?q=%s", url.QueryEscape(query.Query))
		bResp, bBody, bErr := client.Get(ctx, bingURL, &network.RequestOptions{Headers: headers})
		if bErr == nil && bResp.StatusCode == 200 {
			bDoc, _ := goquery.NewDocumentFromReader(bytes.NewReader(bBody))
			if bDoc != nil {
				bDoc.Find("ol#b_results li.b_algo").Each(func(i int, s *goquery.Selection) {
					link := s.Find("h2 a").First()
					title := strings.TrimSpace(link.Text())
					href, _ := link.Attr("href")
					content := strings.TrimSpace(s.Find("p").Text())

					if strings.HasPrefix(href, "https://www.bing.com/ck/a?") {
						if u, err := url.Parse(href); err == nil {
							uVal := u.Query().Get("u")
							if strings.HasPrefix(uVal, "a1") {
								encoded := uVal[2:]
								if pad := len(encoded)%4; pad > 0 {
									encoded += strings.Repeat("=", 4-pad)
								}
								if decoded, err := base64.URLEncoding.DecodeString(encoded); err == nil {
									href = string(decoded)
								}
							}
						}
					}

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

	container.Extend(w.Name(), items)
	return nil
}
