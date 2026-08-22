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

type YahooEngine struct {
	engine.BaseEngine
}

func NewYahooEngine() *YahooEngine {
	return &YahooEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "yahoo",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (y *YahooEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
	}

	searchURL := fmt.Sprintf("https://search.yahoo.com/search?p=%s&ei=UTF-8", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		b := (query.PageNo-1)*10 + 1
		searchURL += fmt.Sprintf("&b=%d", b)
	}

	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("yahoo returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	doc.Find("style, script, svg, span.fc-falcon, span.fz-ms, span.d-ib.fz-14").Remove()

	var items []types.SearchResult
	doc.Find("ol.searchCenterMiddle li, div.algo-sr, div.dd.algo").Each(func(i int, s *goquery.Selection) {
		link := s.Find("h3.title a, a.d-ib, h3 a").First()
		rawHref, _ := link.Attr("href")
		title := strings.TrimSpace(link.Text())
		content := strings.TrimSpace(s.Find("div.compText, p.lh-16, span.fc-smoke, div.abspan, p").First().Text())

		if rawHref == "" || title == "" {
			return
		}

		// Handle Yahoo RU link redirect decoding: RU=.../RK=...
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

	container.Extend(y.Name(), items)
	return nil
}
