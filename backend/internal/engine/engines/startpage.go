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

type StartpageEngine struct {
	engine.BaseEngine
}

func NewStartpageEngine() *StartpageEngine {
	return &StartpageEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "startpage",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (s *StartpageEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	formData := url.Values{
		"query":    {query.Query},
		"cat":      {"web"},
		"cmd":      {"process_search"},
		"language": {"english"},
	}
	if query.PageNo > 1 {
		formData.Set("page", fmt.Sprintf("%d", query.PageNo))
	}

	searchURL := "https://www.startpage.com/sp/search"
	opts := &network.RequestOptions{
		Body:        strings.NewReader(formData.Encode()),
		ContentType: "application/x-www-form-urlencoded",
		Headers: map[string]string{
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
			"Referer":    "https://www.startpage.com/",
		},
	}

	resp, body, err := client.Post(ctx, searchURL, opts)
	var items []types.SearchResult

	if err == nil && resp.StatusCode < 400 {
		doc, dErr := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if dErr == nil {
			// Strip all inline styles and scripts so they don't pollute titles/descriptions
			doc.Find("style, script, svg").Remove()

			doc.Find("div.w-gl__result, div.result").Each(func(i int, sel *goquery.Selection) {
				link := sel.Find("a.w-gl__result-title, a.result-link, h2 a").First()
				href, _ := link.Attr("href")
				title := strings.TrimSpace(link.Text())
				content := strings.TrimSpace(sel.Find("p.w-gl__description, p.result-snippet, p").First().Text())

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

	// Fallback to DuckDuckGo/Bing if Startpage blocks or returns 0
	if len(items) == 0 {
		ddgURL := "https://html.duckduckgo.com/html/"
		dForm := url.Values{"q": {query.Query}, "b": {""}, "kl": {"us-en"}}
		dOpts := &network.RequestOptions{
			Body:        strings.NewReader(dForm.Encode()),
			ContentType: "application/x-www-form-urlencoded",
			Headers:     map[string]string{"Referer": "https://html.duckduckgo.com/"},
		}
		dResp, dBody, dErr := client.Post(ctx, ddgURL, dOpts)
		if dErr == nil && dResp.StatusCode == 200 {
			dDoc, _ := goquery.NewDocumentFromReader(bytes.NewReader(dBody))
			if dDoc != nil {
				dDoc.Find("style, script, svg").Remove()
				dDoc.Find("div.result.results_links").Each(func(i int, sel *goquery.Selection) {
					link := sel.Find("a.result__url").First()
					if link.Length() == 0 {
						link = sel.Find("a.result__a").First()
					}
					rawHref, _ := link.Attr("href")
					title := strings.TrimSpace(sel.Find("a.result__a").Text())
					content := strings.TrimSpace(sel.Find("a.result__snippet").Text())

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

	container.Extend(s.Name(), items)
	return nil
}
