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

type DuckDuckGoEngine struct {
	engine.BaseEngine
}

func NewDuckDuckGoEngine() *DuckDuckGoEngine {
	return &DuckDuckGoEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "duckduckgo",
			EngineCategories: []string{"general", "web"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (d *DuckDuckGoEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	// 1. Concurrently fetch Instant Answers API if on page 1
	if query.PageNo <= 1 {
		go func() {
			apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1", url.QueryEscape(query.Query))
			iaCtx, iaCancel := context.WithTimeout(ctx, 1200*time.Millisecond)
			defer iaCancel()

			if _, body, err := client.Get(iaCtx, apiURL, nil); err == nil {
				var ddgAns struct {
					AbstractText   string `json:"AbstractText"`
					AbstractSource string `json:"AbstractSource"`
					AbstractURL    string `json:"AbstractURL"`
					Answer         string `json:"Answer"`
					Heading        string `json:"Heading"`
					Image          string `json:"Image"`
				}
				if err := json.Unmarshal(body, &ddgAns); err == nil {
					if ddgAns.Answer != "" {
						container.AddAnswer(types.Answer{
							Answer: ddgAns.Answer,
							Engine: d.Name(),
						})
					} else if ddgAns.AbstractText != "" {
						container.AddAnswer(types.Answer{
							Answer: ddgAns.AbstractText,
							URL:    ddgAns.AbstractURL,
							Engine: d.Name(),
						})
					}
					if ddgAns.Heading != "" && ddgAns.AbstractText != "" {
						imgURL := strings.TrimSpace(ddgAns.Image)
						if imgURL != "" {
							if strings.HasPrefix(imgURL, "//") {
								imgURL = "https:" + imgURL
							} else if strings.HasPrefix(imgURL, "/") {
								imgURL = "https://duckduckgo.com" + imgURL
							}
						}

						container.AddInfobox(types.Infobox{
							Infobox: ddgAns.Heading,
							Content: ddgAns.AbstractText,
							ImgSrc:  imgURL,
							Engine:  d.Name(),
							URLs: []types.InfoboxURL{
								{Title: ddgAns.AbstractSource, URL: ddgAns.AbstractURL},
							},
						})
					}
				}
			}
		}()
	}

	// 2. Fetch HTML web search results from DuckDuckGo Lite / HTML
	searchURL := "https://html.duckduckgo.com/html/"
	formData := url.Values{
		"q": {query.Query},
		"b": {""},
		"kl": {"us-en"},
	}

	if query.PageNo > 1 {
		formData.Set("s", fmt.Sprintf("%d", (query.PageNo-1)*30))
	}

	opts := &network.RequestOptions{
		Body:        strings.NewReader(formData.Encode()),
		ContentType: "application/x-www-form-urlencoded",
		Headers: map[string]string{
			"Referer": "https://html.duckduckgo.com/",
		},
	}

	resp, body, err := client.Post(ctx, searchURL, opts)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("duckduckgo returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult
	doc.Find("div.result.results_links").Each(func(i int, s *goquery.Selection) {
		link := s.Find("a.result__url").First()
		if link.Length() == 0 {
			link = s.Find("a.result__a").First()
		}

		rawHref, _ := link.Attr("href")
		title := strings.TrimSpace(s.Find("a.result__a").Text())
		content := strings.TrimSpace(s.Find("a.result__snippet").Text())

		if rawHref == "" || title == "" {
			return
		}

		// DuckDuckGo redirects links via /l/?uddg=...
		finalURL := rawHref
		if strings.Contains(rawHref, "uddg=") {
			if u, err := url.Parse(rawHref); err == nil {
				if uddg := u.Query().Get("uddg"); uddg != "" {
					finalURL = uddg
				}
			}
		}

		items = append(items, types.SearchResult{
			URL:      finalURL,
			Title:    title,
			Content:  content,
			Category: "general",
			Template: "default.html",
		})
	})

	container.Extend(d.Name(), items)
	return nil
}
