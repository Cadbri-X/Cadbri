package results

import (
	"testing"

	"cadbri/internal/types"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://www.example.com/path/?utm_source=google&utm_medium=cpc#section",
			expected: "https://www.example.com/path",
		},
		{
			input:    "http://Example.com/page/",
			expected: "http://example.com/page",
		},
		{
			input:    "https://example.com",
			expected: "https://example.com/",
		},
	}

	for _, tt := range tests {
		got := NormalizeURL(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeURL(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestResultContainerDeduplicationAndScoring(t *testing.T) {
	weights := map[string]float64{
		"google":     1.0,
		"duckduckgo": 1.0,
		"bing":       1.0,
	}
	rc := NewResultContainer(weights)

	// Engine 1 results
	rc.Extend("duckduckgo", []types.SearchResult{
		{URL: "https://www.facebook.com/", Title: "Facebook", Content: "Log in or sign up"},
		{URL: "https://en.wikipedia.org/wiki/Facebook", Title: "Facebook - Wikipedia", Content: "Online social media service"},
	})

	// Engine 2 results (has duplicate Facebook, higher up)
	rc.Extend("bing", []types.SearchResult{
		{URL: "https://www.facebook.com/?utm_source=bing", Title: "Facebook", Content: "Connect and share with friends and family on Facebook"},
		{URL: "https://www.instagram.com/", Title: "Instagram", Content: "Photo and video sharing"},
	})

	// Engine 3 results (Facebook also on Google)
	rc.Extend("google", []types.SearchResult{
		{URL: "https://www.facebook.com/#top", Title: "Facebook", Content: "Social networking service"},
	})

	results := rc.GetOrderedResults()

	// facebook.com appeared in all 3 engines -> should be #1 ranked with highest score
	if len(results) != 3 {
		t.Fatalf("expected 3 deduplicated results, got %d", len(results))
	}

	top := results[0]
	if top.URL != "https://www.facebook.com/" {
		t.Errorf("expected top result to be Facebook, got %s", top.URL)
	}

	if len(top.Engines) != 3 {
		t.Errorf("expected Facebook to have 3 engines, got %v", top.Engines)
	}

	if top.Score <= 1.0 {
		t.Errorf("expected boosted score for multi-engine result, got %f", top.Score)
	}
}
