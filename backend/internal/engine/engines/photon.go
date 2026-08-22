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

type PhotonEngine struct {
	engine.BaseEngine
}

func NewPhotonEngine() *PhotonEngine {
	return &PhotonEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "photon",
			EngineCategories: []string{"map"},
			EngineWeight:     1.0,
			EngineTimeout:    3 * time.Second,
			CanPage:          false,
		},
	}
}

type photonGeoJSON struct {
	Features []struct {
		Geometry struct {
			Coordinates []float64 `json:"coordinates"` // [lon, lat]
		} `json:"geometry"`
		Properties struct {
			Name     string `json:"name"`
			Country  string `json:"country"`
			City     string `json:"city"`
			State    string `json:"state"`
			Street   string `json:"street"`
			Postcode string `json:"postcode"`
			Type     string `json:"type"`
		} `json:"properties"`
	} `json:"features"`
}

func (p *PhotonEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	apiURL := fmt.Sprintf("https://photon.komoot.io/api/?q=%s&limit=10", url.QueryEscape(query.Query))
	_, body, err := client.Get(ctx, apiURL, nil)
	if err != nil {
		return err
	}

	var data photonGeoJSON
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	var items []types.SearchResult
	for _, f := range data.Features {
		if len(f.Geometry.Coordinates) < 2 {
			continue
		}
		lon := f.Geometry.Coordinates[0]
		lat := f.Geometry.Coordinates[1]

		props := f.Properties
		var parts []string
		if props.Street != "" {
			parts = append(parts, props.Street)
		}
		if props.City != "" {
			parts = append(parts, props.City)
		}
		if props.State != "" {
			parts = append(parts, props.State)
		}
		if props.Country != "" {
			parts = append(parts, props.Country)
		}
		addr := strings.Join(parts, ", ")

		title := props.Name
		if title == "" {
			title = addr
		}

		mapURL := fmt.Sprintf("https://www.openstreetmap.org/?mlat=%f&mlon=%f#map=15/%f/%f", lat, lon, lat, lon)
		desc := fmt.Sprintf("[Coordinates: %f, %f] Address: %s", lat, lon, addr)

		items = append(items, types.SearchResult{
			URL:      mapURL,
			Title:    title,
			Content:  desc,
			Category: "map",
			Template: "map.html",
		})
	}

	container.Extend(p.Name(), items)
	return nil
}
