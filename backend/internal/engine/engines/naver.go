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

type NaverEngine struct {
	engine.BaseEngine
}

func NewNaverEngine() *NaverEngine {
	return &NaverEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "naver",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (n *NaverEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://search.naver.com/search.naver?where=nexearch&query=%s", url.QueryEscape(query.Query))
	if query.PageNo > 1 {
		start := (query.PageNo-1)*15 + 1
		searchURL += fmt.Sprintf("&start=%d", start)
	}

	headers := map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "ko-KR,ko;q=0.9,en;q=0.8",
	}

	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("naver returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("div.total_wrap, li.bx, div.bx_inner, div.api_subject_bx").Each(func(i int, s *goquery.Selection) {
		link := s.Find("a.total_tit, a.link_tit, a.news_tit, a.tit").First()
		href, _ := link.Attr("href")
		title := strings.TrimSpace(link.Text())
		content := strings.TrimSpace(s.Find("div.total_dsc, div.dsc_wrap, div.news_dsc, div.dsc_txt").First().Text())

		if href != "" && title != "" && !strings.HasPrefix(href, "javascript:") {
			items = append(items, types.SearchResult{
				URL:      href,
				Title:    title,
				Content:  content,
				Category: "general",
				Template: "default.html",
			})
		}
	})

	container.Extend(n.Name(), items)
	return nil
}
