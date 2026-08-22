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

type MailRuEngine struct {
	engine.BaseEngine
}

func NewMailRuEngine() *MailRuEngine {
	return &MailRuEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "mailru",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     0.9,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (m *MailRuEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
	}

	searchURL := fmt.Sprintf("https://go.mail.ru/search?q=%s", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		sf := (query.PageNo - 1) * 10
		searchURL += fmt.Sprintf("&sf=%d", sf)
	}

	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	var items []types.SearchResult

	if err == nil && resp.StatusCode == 200 {
		doc, dErr := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if dErr == nil {
			doc.Find("style, script, svg").Remove()
			doc.Find("li.Snippet, div.Snippet, div.result__item, div.results-item, article, li.results__item").Each(func(i int, s *goquery.Selection) {
				link := s.Find("a.Snippet-title, h3 a, a.link, a.result__title-link, a").First()
				href, _ := link.Attr("href")
				title := strings.TrimSpace(link.Text())
				content := strings.TrimSpace(s.Find("div.Snippet-text, div.Snippet-desc, span.Snippet-text, p.result__description, p").First().Text())

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

	// Fallback to Yandex/DuckDuckGo if Mail.ru is rate limited or empty
	if len(items) == 0 {
		yandexURL := fmt.Sprintf("https://yandex.com/search/?text=%s", url.QueryEscape(query.Query))
		yResp, yBody, yErr := client.Get(ctx, yandexURL, &network.RequestOptions{Headers: headers})
		if yErr == nil && yResp.StatusCode == 200 {
			yDoc, _ := goquery.NewDocumentFromReader(bytes.NewReader(yBody))
			if yDoc != nil {
				yDoc.Find("style, script, svg").Remove()
				yDoc.Find("li.serp-item, div.serp-item").Each(func(i int, s *goquery.Selection) {
					link := s.Find("h2 a, a.organic__url, a.link").First()
					href, _ := link.Attr("href")
					title := strings.TrimSpace(link.Text())
					content := strings.TrimSpace(s.Find("div.OrganicTextContentSpan, div.organic__content-wrapper, div.text-container, p").First().Text())

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

	container.Extend(m.Name(), items)
	return nil
}
