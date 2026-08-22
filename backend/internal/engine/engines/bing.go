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

type BingEngine struct {
	engine.BaseEngine
}

func NewBingEngine() *BingEngine {
	return &BingEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "bing",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (b *BingEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	safeMap := map[int]string{0: "off", 1: "moderate", 2: "strict"}
	safeVal := safeMap[query.SafeSearch]
	if safeVal == "" {
		safeVal = "off"
	}

	searchURL := fmt.Sprintf("https://www.bing.com/search?q=%s&adlt=%s", url.QueryEscape(query.Query), safeVal)
	if query.PageNo > 1 {
		first := (query.PageNo-1)*10 + 1
		searchURL += fmt.Sprintf("&first=%d", first)
	}

	headers := map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
	}

	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("bing returned HTTP status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("ol#b_results li.b_algo").Each(func(i int, s *goquery.Selection) {
		link := s.Find("h2 a").First()
		title := strings.TrimSpace(link.Text())
		href, _ := link.Attr("href")

		if href == "" || title == "" {
			return
		}

		// Handle Bing redirect URL (https://www.bing.com/ck/a?...)
		if strings.HasPrefix(href, "https://www.bing.com/ck/a?") {
			if u, err := url.Parse(href); err == nil {
				uVal := u.Query().Get("u")
				if strings.HasPrefix(uVal, "a1") {
					encoded := uVal[2:]
					if pad := len(encoded) % 4; pad > 0 {
						encoded += strings.Repeat("=", 4-pad)
					}
					if decoded, err := base64.URLEncoding.DecodeString(encoded); err == nil {
						href = string(decoded)
					}
				}
			}
		}

		// Clean up snippet content
		s.Find("p .algoSlug_icon").Remove()
		content := strings.TrimSpace(s.Find("p").Text())

		items = append(items, types.SearchResult{
			URL:      href,
			Title:    title,
			Content:  content,
			Category: "general",
			Template: "default.html",
		})
	})

	// Total result count
	if countText := doc.Find("span.sb_count").Text(); countText != "" {
		digits := ""
		for _, r := range countText {
			if r >= '0' && r <= '9' {
				digits += string(r)
			}
		}
		if num, err := strconv.ParseInt(digits, 10, 64); err == nil && num > 0 {
			container.SetNumberOfResults(num)
		}
	}

	container.Extend(b.Name(), items)
	return nil
}
