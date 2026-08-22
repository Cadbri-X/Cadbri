package engines

import (
	"bytes"
	"context"
	"encoding/json"
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

type MarginaliaEngine struct {
	engine.BaseEngine
}

func NewMarginaliaEngine() *MarginaliaEngine {
	return &MarginaliaEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "marginalia",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

type marginaliaJSON struct {
	Results []struct {
		URL         string `json:"url"`
		Title       string `json:"title"`
		Description string `json:"description"`
	} `json:"results"`
}

func (m *MarginaliaEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	}

	searchURL := fmt.Sprintf("https://search.marginalia.nu/search?query=%s", url.QueryEscape(query.Query))
	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	var items []types.SearchResult

	if err == nil && resp.StatusCode == 200 {
		doc, dErr := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if dErr == nil {
			doc.Find("section.search-result, div.search-result").Each(func(i int, s *goquery.Selection) {
				link := s.Find("h2.title a, a.url, a").First()
				href, _ := link.Attr("href")
				title := strings.TrimSpace(link.Text())
				content := strings.TrimSpace(s.Find("p.description, div.description").Text())

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

	// Try API if HTML returned 0 results
	if len(items) == 0 {
		apiURL := fmt.Sprintf("https://api.marginalia.nu/public/search/%s", url.QueryEscape(query.Query))
		aResp, aBody, aErr := client.Get(ctx, apiURL, &network.RequestOptions{
			Headers: map[string]string{"Accept": "application/json"},
		})
		if aErr == nil && aResp.StatusCode == 200 {
			var d marginaliaJSON
			if json.Unmarshal(aBody, &d) == nil {
				for _, r := range d.Results {
					if r.URL != "" && r.Title != "" {
						items = append(items, types.SearchResult{
							URL:      r.URL,
							Title:    r.Title,
							Content:  r.Description,
							Category: "general",
							Template: "default.html",
						})
					}
				}
			}
		}
	}

	// Fallback to Wiby / DDG if non-indexed
	if len(items) == 0 {
		wibyURL := fmt.Sprintf("https://wiby.me/?q=%s", url.QueryEscape(query.Query))
		wResp, wBody, wErr := client.Get(ctx, wibyURL, &network.RequestOptions{Headers: headers})
		if wErr == nil && wResp.StatusCode == 200 {
			wDoc, _ := goquery.NewDocumentFromReader(bytes.NewReader(wBody))
			if wDoc != nil {
				wDoc.Find("table tr, div.result, p").Each(func(i int, s *goquery.Selection) {
					link := s.Find("a").First()
					href, _ := link.Attr("href")
					title := strings.TrimSpace(link.Text())
					content := strings.TrimSpace(s.Find("small, span, p").Text())
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
