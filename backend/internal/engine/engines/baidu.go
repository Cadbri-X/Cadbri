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

type BaiduEngine struct {
	engine.BaseEngine
}

func NewBaiduEngine() *BaiduEngine {
	return &BaiduEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "baidu",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (b *BaiduEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://www.baidu.com/s?wd=%s&ie=utf-8", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		pn := (query.PageNo - 1) * 10
		searchURL += fmt.Sprintf("&pn=%d", pn)
	}

	headers := map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
	}

	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("baidu returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("div.result.c-container, div.c-container").Each(func(i int, s *goquery.Selection) {
		link := s.Find("h3 a, a.c-title-text").First()
		href, _ := link.Attr("href")
		title := strings.TrimSpace(link.Text())
		content := strings.TrimSpace(s.Find(".c-abstract, .content-right_8Zs40, .c-font-normal, div.cos-row").First().Text())

		if href != "" && title != "" {
			items = append(items, types.SearchResult{
				URL:      href,
				Title:    title,
				Content:  content,
				Category: "general",
				Template: "default.html",
			})
		}
	})

	container.Extend(b.Name(), items)
	return nil
}
