package engines

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"cadbri/internal/engine"
	"cadbri/internal/network"
	"cadbri/internal/results"
	"cadbri/internal/types"
)

type AppleMapsEngine struct {
	engine.BaseEngine
}

func NewAppleMapsEngine() *AppleMapsEngine {
	return &AppleMapsEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "apple_maps",
			EngineCategories: []string{"map"},
			EngineWeight:     1.0,
			EngineTimeout:    2 * time.Second,
			CanPage:          false,
		},
	}
}

func (a *AppleMapsEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	appleURL := fmt.Sprintf("https://maps.apple.com/?q=%s", url.QueryEscape(query.Query))

	items := []types.SearchResult{
		{
			URL:      appleURL,
			Title:    fmt.Sprintf("Apple Maps: %s", query.Query),
			Content:  fmt.Sprintf("View '%s' in Apple Maps navigation and satellite imagery", query.Query),
			Category: "map",
			Template: "map.html",
		},
	}

	container.Extend(a.Name(), items)
	return nil
}
