package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"cadbri/internal/engine"
	"cadbri/internal/network"
	"cadbri/internal/results"
	"cadbri/internal/types"
)

type PubMedEngine struct {
	engine.BaseEngine
}

func NewPubMedEngine() *PubMedEngine {
	return &PubMedEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "pubmed",
			EngineCategories: []string{"science", "academic"},
			EngineWeight:     1.1,
			EngineTimeout:    4 * time.Second,
			CanPage:          true,
		},
	}
}

type esearchResponse struct {
	ESearchResult struct {
		Count    string   `json:"count"`
		IDList   []string `json:"idlist"`
	} `json:"esearchresult"`
}

type esummaryResponse struct {
	Result map[string]interface{} `json:"result"`
}

func (p *PubMedEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	retStart := 0
	if query.PageNo > 1 {
		retStart = (query.PageNo - 1) * 10
	}

	searchURL := fmt.Sprintf(
		"https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi?db=pubmed&term=%s&retstart=%d&retmax=10&retmode=json",
		url.QueryEscape(query.Query), retStart,
	)

	_, body, err := client.Get(ctx, searchURL, nil)
	if err != nil {
		return err
	}

	var searchData esearchResponse
	if err := json.Unmarshal(body, &searchData); err != nil {
		return err
	}

	if len(searchData.ESearchResult.IDList) == 0 {
		return nil
	}

	ids := strings.Join(searchData.ESearchResult.IDList, ",")
	summaryURL := fmt.Sprintf(
		"https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esummary.fcgi?db=pubmed&id=%s&retmode=json",
		ids,
	)

	_, sumBody, err := client.Get(ctx, summaryURL, nil)
	if err != nil {
		return err
	}

	var sumData esummaryResponse
	if err := json.Unmarshal(sumBody, &sumData); err != nil {
		return err
	}

	var items []types.SearchResult
	for _, id := range searchData.ESearchResult.IDList {
		rawDoc, found := sumData.Result[id]
		if !found {
			continue
		}

		docMap, ok := rawDoc.(map[string]interface{})
		if !ok {
			continue
		}

		title, _ := docMap["title"].(string)
		pubdate, _ := docMap["pubdate"].(string)
		source, _ := docMap["source"].(string)

		paperURL := fmt.Sprintf("https://pubmed.ncbi.nlm.nih.gov/%s/", id)
		desc := fmt.Sprintf("[%s | %s]", source, pubdate)

		items = append(items, types.SearchResult{
			URL:      paperURL,
			Title:    title,
			Content:  desc,
			Category: "science",
			Template: "paper.html",
			PubDate:  pubdate,
		})
	}

	container.Extend(p.Name(), items)
	return nil
}
