package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"cadbri/internal/network"
)

// Autocompleter fetches instant search suggestions from upstream autocomplete providers.
type Autocompleter struct {
	client *network.Client
}

// NewAutocompleter creates a new autocompleter instance.
func NewAutocompleter(client *network.Client) *Autocompleter {
	return &Autocompleter{client: client}
}

// Complete returns a list of query completions for a given prefix.
func (a *Autocompleter) Complete(ctx context.Context, prefix, backend string) []string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return nil
	}

	backend = strings.ToLower(backend)
	switch backend {
	case "google":
		return a.completeGoogle(ctx, prefix)
	case "bing":
		return a.completeBing(ctx, prefix)
	case "wikipedia":
		return a.completeWikipedia(ctx, prefix)
	default:
		// Default to DuckDuckGo
		return a.completeDuckDuckGo(ctx, prefix)
	}
}

func (a *Autocompleter) completeDuckDuckGo(ctx context.Context, prefix string) []string {
	apiURL := fmt.Sprintf("https://duckduckgo.com/ac/?q=%s&type=list", url.QueryEscape(prefix))
	_, body, err := a.client.Get(ctx, apiURL, nil)
	if err != nil {
		return a.completeGoogle(ctx, prefix) // Fallback
	}

	// Returns [query, [suggestions...]]
	var data []interface{}
	if err := json.Unmarshal(body, &data); err != nil || len(data) < 2 {
		return nil
	}

	sugList, ok := data[1].([]interface{})
	if !ok {
		return nil
	}

	var results []string
	for _, s := range sugList {
		if str, ok := s.(string); ok && str != "" {
			results = append(results, str)
		}
	}
	return results
}

func (a *Autocompleter) completeGoogle(ctx context.Context, prefix string) []string {
	apiURL := fmt.Sprintf("https://suggestqueries.google.com/complete/search?client=chrome&q=%s", url.QueryEscape(prefix))
	_, body, err := a.client.Get(ctx, apiURL, nil)
	if err != nil {
		return nil
	}

	var data []interface{}
	if err := json.Unmarshal(body, &data); err != nil || len(data) < 2 {
		return nil
	}

	sugList, ok := data[1].([]interface{})
	if !ok {
		return nil
	}

	var results []string
	for _, s := range sugList {
		if str, ok := s.(string); ok && str != "" {
			results = append(results, str)
		}
	}
	return results
}

func (a *Autocompleter) completeBing(ctx context.Context, prefix string) []string {
	apiURL := fmt.Sprintf("https://api.bing.com/osjson.aspx?query=%s", url.QueryEscape(prefix))
	_, body, err := a.client.Get(ctx, apiURL, nil)
	if err != nil {
		return nil
	}

	var data []interface{}
	if err := json.Unmarshal(body, &data); err != nil || len(data) < 2 {
		return nil
	}

	sugList, ok := data[1].([]interface{})
	if !ok {
		return nil
	}

	var results []string
	for _, s := range sugList {
		if str, ok := s.(string); ok && str != "" {
			results = append(results, str)
		}
	}
	return results
}

func (a *Autocompleter) completeWikipedia(ctx context.Context, prefix string) []string {
	apiURL := fmt.Sprintf("https://en.wikipedia.org/w/api.php?action=opensearch&search=%s&limit=10&format=json", url.QueryEscape(prefix))
	_, body, err := a.client.Get(ctx, apiURL, nil)
	if err != nil {
		return nil
	}

	var data []interface{}
	if err := json.Unmarshal(body, &data); err != nil || len(data) < 2 {
		return nil
	}

	sugList, ok := data[1].([]interface{})
	if !ok {
		return nil
	}

	var results []string
	for _, s := range sugList {
		if str, ok := s.(string); ok && str != "" {
			results = append(results, str)
		}
	}
	return results
}
