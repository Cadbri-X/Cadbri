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

type AOLEngine struct {
	engine.BaseEngine
}

func NewAOLEngine() *AOLEngine {
	return &AOLEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "aol",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (a *AOLEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
	}

	searchURL := fmt.Sprintf("https://search.aol.com/aol/search?q=%s&ei=UTF-8", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		b := (query.PageNo-1)*10 + 1
		searchURL += fmt.Sprintf("&b=%d", b)
	}

	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	var items []types.SearchResult

	if err == nil && resp.StatusCode == 200 {
		doc, dErr := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if dErr == nil {
			doc.Find("style, script, svg, span.fc-falcon, span.fz-ms, span.d-ib.fz-14").Remove()
			doc.Find("ol.searchCenterMiddle li, div.algo-sr, div.dd.algo, div.compText").Each(func(i int, s *goquery.Selection) {
				link := s.Find("h3.title a, a.d-ib, h3 a").First()
				rawHref, _ := link.Attr("href")
				title := strings.TrimSpace(link.Text())
				content := strings.TrimSpace(s.Find("div.compText, p.lh-16, span.fc-smoke, div.abspan, p").First().Text())

				if rawHref == "" || title == "" {
					return
				}

				// Handle Yahoo/AOL RU redirect decoding
				finalURL := rawHref
				if strings.Contains(rawHref, "/RU=") {
					parts := strings.Split(rawHref, "/RU=")
					if len(parts) > 1 {
						sub := strings.Split(parts[1], "/RK=")[0]
						if decoded, err := url.QueryUnescape(sub); err == nil {
							finalURL = decoded
						}
					}
				}

				if strings.HasPrefix(finalURL, "http") {
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

	// Fallback to Yahoo/DuckDuckGo format if AOL is blocked
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
				dDoc.Find("style, script, svg").Remove()
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

	container.Extend(a.Name(), items)
	return nil
}
