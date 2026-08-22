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

type SoundCloudEngine struct {
	engine.BaseEngine
}

func NewSoundCloudEngine() *SoundCloudEngine {
	return &SoundCloudEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "soundcloud",
			EngineCategories: []string{"music", "media"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (s *SoundCloudEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://soundcloud.com/search/sounds?q=%s", url.QueryEscape(query.Query))
	resp, body, err := client.Get(ctx, searchURL, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("soundcloud returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("li, div.sound, article").Each(func(i int, sel *goquery.Selection) {
		link := sel.Find("a[itemprop='url'], h2 a").First()
		href, _ := link.Attr("href")
		title := strings.TrimSpace(link.Text())

		if href != "" && title != "" && !strings.Contains(href, "/search?") {
			if strings.HasPrefix(href, "/") {
				href = "https://soundcloud.com" + href
			}
			items = append(items, types.SearchResult{
				URL:      href,
				Title:    title,
				Content:  "SoundCloud Audio Track",
				Category: "music",
				Template: "audio.html",
			})
		}
	})

	container.Extend(s.Name(), items)
	return nil
}
