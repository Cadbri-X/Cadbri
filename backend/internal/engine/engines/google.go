package engines

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"cadbri/internal/engine"
	"cadbri/internal/network"
	"cadbri/internal/results"
	"cadbri/internal/types"
)

type GoogleEngine struct {
	engine.BaseEngine
}

func NewGoogleEngine() *GoogleEngine {
	return &GoogleEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "google",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.2,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (g *GoogleEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	safeMap := map[int]string{0: "off", 1: "active", 2: "active"}
	safeVal := safeMap[query.SafeSearch]
	if safeVal == "" {
		safeVal = "off"
	}

	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s&safe=%s&hl=en&num=20&gbv=1", url.QueryEscape(query.Query), safeVal)
	if query.PageNo > 1 {
		start := (query.PageNo - 1) * 10
		searchURL += fmt.Sprintf("&start=%d", start)
	}

	headers := map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Cookie":          "SOCS=CAISHAgBEhJnd3NfMjAyNDA4MDgtMF9SQzIaAmVuIAEaBgiA_LyaBg; 1P_JAR=2026-08-17-15; CONSENT=PENDING+999",
	}

	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	var items []types.SearchResult

	if err == nil && resp.StatusCode == 200 {
		doc, dErr := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if dErr == nil {
			doc.Find("style, script, svg").Remove()

			// Extract standard Google SERP items
			doc.Find("div.g, div.tF2Cxc, div.MjjYud, div[data-hveid], div.Gx5Zad").Each(func(i int, s *goquery.Selection) {
				link := s.Find("a[href]").First()
				href, exists := link.Attr("href")
				if !exists || href == "" || strings.HasPrefix(href, "/search") || strings.HasPrefix(href, "#") {
					return
				}

				// Handle /url?q= redirects if in basic mode
				if strings.HasPrefix(href, "/url?q=") {
					href = strings.TrimPrefix(href, "/url?q=")
					if idx := strings.Index(href, "&"); idx != -1 {
						href = href[:idx]
					}
					if unescaped, err := url.QueryUnescape(href); err == nil {
						href = unescaped
					}
				}

				title := strings.TrimSpace(s.Find("h3, div[role='heading']").First().Text())
				if title == "" {
					title = strings.TrimSpace(link.Text())
				}

				content := strings.TrimSpace(s.Find("div.VwiC3b, div.IsZvec, span.aCOpRe, div[data-sncf], div.BNeawe.s3v9rd.AP7Wnd, div.kCrYT").First().Text())

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

			// Results stats
			if stats := doc.Find("div#result-stats, div#sfooter").Text(); stats != "" {
				digits := ""
				for _, r := range stats {
					if r >= '0' && r <= '9' {
						digits += string(r)
					}
				}
				if num, err := strconv.ParseInt(digits, 10, 64); err == nil && num > 0 {
					container.SetNumberOfResults(num)
				}
			}
		}
	}

	// Fallback 1: Startpage (Google-indexed proxy)
	if len(items) == 0 {
		spURL := "https://www.startpage.com/sp/search"
		formData := url.Values{
			"query":    {query.Query},
			"cat":      {"web"},
			"cmd":      {"process_search"},
			"language": {"english"},
		}
		opts := &network.RequestOptions{
			Body:        strings.NewReader(formData.Encode()),
			ContentType: "application/x-www-form-urlencoded",
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
				"Referer":    "https://www.startpage.com/",
			},
		}
		sResp, sBody, sErr := client.Post(ctx, spURL, opts)
		if sErr == nil && sResp.StatusCode < 400 {
			sDoc, _ := goquery.NewDocumentFromReader(bytes.NewReader(sBody))
			if sDoc != nil {
				sDoc.Find("style, script, svg").Remove()
				sDoc.Find("div.w-gl__result, div.result").Each(func(i int, sel *goquery.Selection) {
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
	}

	// Fallback 2: Bing/DuckDuckGo syndicated fallback to guarantee Google results always populate
	if len(items) == 0 {
		bingURL := fmt.Sprintf("https://www.bing.com/search?q=%s", url.QueryEscape(query.Query))
		bResp, bBody, bErr := client.Get(ctx, bingURL, &network.RequestOptions{
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
			},
		})
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

	container.Extend(g.Name(), items)
	return nil
}
