package engines

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
	"time"

	"cadbri/internal/engine"
	"cadbri/internal/network"
	"cadbri/internal/results"
	"cadbri/internal/types"
)

type GoogleNewsEngine struct {
	engine.BaseEngine
}

func NewGoogleNewsEngine() *GoogleNewsEngine {
	return &GoogleNewsEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "google_news",
			EngineCategories: []string{"news"},
			EngineWeight:     1.1,
			EngineTimeout:    3 * time.Second,
			CanPage:          false,
		},
	}
}

type googleNewsRSS struct {
	Channel struct {
		Title string `xml:"title"`
		Items []struct {
			Title   string `xml:"title"`
			Link    string `xml:"link"`
			PubDate string `xml:"pubDate"`
			Source  struct {
				URL  string `xml:"url,attr"`
				Name string `xml:",chardata"`
			} `xml:"source"`
		} `xml:"item"`
	} `xml:"channel"`
}

func (g *GoogleNewsEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	rssURL := fmt.Sprintf(
		"https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en",
		url.QueryEscape(query.Query),
	)

	_, body, err := client.Get(ctx, rssURL, nil)
	if err != nil {
		return err
	}

	var rss googleNewsRSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return err
	}

	var items []types.SearchResult
	for _, it := range rss.Channel.Items {
		sourceName := it.Source.Name
		title := it.Title
		if sourceName != "" && strings.HasSuffix(title, " - "+sourceName) {
			title = strings.TrimSuffix(title, " - "+sourceName)
		}

		desc := fmt.Sprintf("[%s | %s]", sourceName, it.PubDate)

		items = append(items, types.SearchResult{
			URL:      it.Link,
			Title:    title,
			Content:  desc,
			Category: "news",
			Template: "news.html",
			Author:   sourceName,
			PubDate:  it.PubDate,
		})
	}

	container.Extend(g.Name(), items)
	return nil
}
