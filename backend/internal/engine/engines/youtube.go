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

type YouTubeEngine struct {
	engine.BaseEngine
}

func NewYouTubeEngine() *YouTubeEngine {
	return &YouTubeEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "youtube",
			EngineCategories: []string{"videos", "media"},
			EngineWeight:     1.2,
			EngineTimeout:    3 * time.Second,
			CanPage:          true,
		},
	}
}

func (y *YouTubeEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	searchURL := fmt.Sprintf("https://www.youtube.com/results?search_query=%s", url.QueryEscape(query.Query))
	headers := map[string]string{
		"Accept-Language": "en-US,en;q=0.9",
	}

	resp, body, err := client.Get(ctx, searchURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("youtube returned HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var items []types.SearchResult

	// Extract ytInitialData script JSON
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "var ytInitialData =") || strings.Contains(text, "window[\"ytInitialData\"] =") {
			startIdx := strings.Index(text, "{")
			endIdx := strings.LastIndex(text, "}")
			if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
				jsonData := text[startIdx : endIdx+1]
				var root map[string]interface{}
				if err := json.Unmarshal([]byte(jsonData), &root); err == nil {
					items = append(items, parseYouTubeInitialData(root)...)
				}
			}
		}
	})

	container.Extend(y.Name(), items)
	return nil
}

func parseYouTubeInitialData(data map[string]interface{}) []types.SearchResult {
	var results []types.SearchResult

	defer func() {
		// Recover from any nested JSON navigation panic safely
		recover()
	}()

	contents, ok := data["contents"].(map[string]interface{})
	if !ok {
		return results
	}
	twoCol, ok := contents["twoColumnSearchResultsRenderer"].(map[string]interface{})
	if !ok {
		return results
	}
	primary, ok := twoCol["primaryContents"].(map[string]interface{})
	if !ok {
		return results
	}
	sectionList, ok := primary["sectionListRenderer"].(map[string]interface{})
	if !ok {
		return results
	}
	contentArray, ok := sectionList["contents"].([]interface{})
	if !ok {
		return results
	}

	for _, c := range contentArray {
		itemSection, ok := c.(map[string]interface{})["itemSectionRenderer"].(map[string]interface{})
		if !ok {
			continue
		}
		items, ok := itemSection["contents"].([]interface{})
		if !ok {
			continue
		}

		for _, it := range items {
			video, ok := it.(map[string]interface{})["videoRenderer"].(map[string]interface{})
			if !ok {
				continue
			}

			videoID, _ := video["videoId"].(string)
			if videoID == "" {
				continue
			}

			titleRuns, _ := video["title"].(map[string]interface{})["runs"].([]interface{})
			title := ""
			if len(titleRuns) > 0 {
				title, _ = titleRuns[0].(map[string]interface{})["text"].(string)
			}

			descRuns, _ := video["detailedMetadataSnippets"].([]interface{})
			desc := ""
			if len(descRuns) > 0 {
				snippets, _ := descRuns[0].(map[string]interface{})["snippetText"].(map[string]interface{})["runs"].([]interface{})
				if len(snippets) > 0 {
					desc, _ = snippets[0].(map[string]interface{})["text"].(string)
				}
			}

			thumbURL := fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", videoID)
			videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
			embedURL := fmt.Sprintf("https://www.youtube-nocookie.com/embed/%s", videoID)

			results = append(results, types.SearchResult{
				URL:       videoURL,
				Title:     title,
				Content:   desc,
				Category:  "videos",
				Template:  "videos.html",
				Thumbnail: thumbURL,
				IframeSrc: embedURL,
			})
		}
	}

	return results
}
