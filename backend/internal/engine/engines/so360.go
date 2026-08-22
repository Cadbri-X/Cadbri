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

type So360Engine struct {
	engine.BaseEngine
}

func NewSo360Engine() *So360Engine {
	return &So360Engine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "so360",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (s *So360Engine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
	}

	searchURL := fmt.Sprintf("https://www.so.com/s?q=%s", url.QueryEscape(query.Query))
	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	var items []types.SearchResult

	if err == nil && resp.StatusCode == 200 {
		doc, dErr := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if dErr == nil {
			doc.Find("li.res-list").Each(func(i int, sel *goquery.Selection) {
				link := sel.Find("h3.res-title a, h3 a").First()
				href, _ := link.Attr("href")
				if dataURL, exists := link.Attr("data-url"); exists && strings.HasPrefix(dataURL, "http") {
					href = dataURL
				}
				title := strings.TrimSpace(link.Text())
				content := strings.TrimSpace(sel.Find("p.res-desc, div.res-rich, p").First().Text())

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

	// Fallback to Baidu if 360 search connection is blocked overseas
	if len(items) == 0 {
		baiduURL := fmt.Sprintf("https://www.baidu.com/s?wd=%s", url.QueryEscape(query.Query))
		bResp, bBody, bErr := client.Get(ctx, baiduURL, &network.RequestOptions{Headers: headers})
		if bErr == nil && bResp.StatusCode == 200 {
			bDoc, _ := goquery.NewDocumentFromReader(bytes.NewReader(bBody))
			if bDoc != nil {
				bDoc.Find("div.result, div.c-container").Each(func(i int, sel *goquery.Selection) {
					link := sel.Find("h3 a").First()
					title := strings.TrimSpace(link.Text())
					href, _ := link.Attr("href")
					content := strings.TrimSpace(sel.Find("div.c-abstract, div.content-right_8Zs40, div.c-span-last p").First().Text())

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

	container.Extend(s.Name(), items)
	return nil
}
