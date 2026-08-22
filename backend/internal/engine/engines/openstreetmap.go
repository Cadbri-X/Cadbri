package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"cadbri/internal/engine"
	"cadbri/internal/network"
	"cadbri/internal/results"
	"cadbri/internal/types"
)

type OpenStreetMapEngine struct {
	engine.BaseEngine
}

func NewOpenStreetMapEngine() *OpenStreetMapEngine {
	return &OpenStreetMapEngine{
		BaseEngine: engine.BaseEngine{
			EngineName:       "openstreetmap",
			EngineCategories: []string{"map"},
			EngineWeight:     1.1,
			EngineTimeout:    3 * time.Second,
			CanPage:          false,
		},
	}
}

type osmItem struct {
	PlaceID     int64   `json:"place_id"`
	Lat         string  `json:"lat"`
	Lon         string  `json:"lon"`
	DisplayName string  `json:"display_name"`
	Class       string  `json:"class"`
	Type        string  `json:"type"`
	Importance  float64 `json:"importance"`
}

func (o *OpenStreetMapEngine) Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error {
	apiURL := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/search?q=%s&format=json&addressdetails=1&limit=10",
		url.QueryEscape(query.Query),
	)

	headers := map[string]string{
		"User-Agent": "Cadbri-Go-Map-Aggregator/1.0",
	}

	_, body, err := client.Get(ctx, apiURL, &network.RequestOptions{Headers: headers})
	if err != nil {
		return err
	}

	var places []osmItem
	if err := json.Unmarshal(body, &places); err != nil {
		return err
	}

	var items []types.SearchResult
	for i, it := range places {
		osmURL := fmt.Sprintf("https://www.openstreetmap.org/?mlat=%s&mlon=%s#map=15/%s/%s", it.Lat, it.Lon, it.Lat, it.Lon)
		desc := fmt.Sprintf("[Coordinates: %s, %s | Type: %s/%s]", it.Lat, it.Lon, it.Class, it.Type)

		items = append(items, types.SearchResult{
			URL:      osmURL,
			Title:    it.DisplayName,
			Content:  desc,
			Category: "map",
			Template: "map.html",
		})

		if i == 0 && query.PageNo <= 1 {
			container.AddAnswer(types.Answer{
				Answer: fmt.Sprintf("Location: %s (Lat: %s, Lon: %s)", it.DisplayName, it.Lat, it.Lon),
				URL:    osmURL,
				Engine: o.Name(),
			})
			container.AddInfobox(types.Infobox{
				Infobox: it.DisplayName,
				ID:      osmURL,
				Content: desc,
				Engine:  o.Name(),
				URLs: []types.InfoboxURL{
					{Title: "OpenStreetMap", URL: osmURL},
				},
				Attributes: []types.InfoboxAttribute{
					{Label: "Latitude", Value: it.Lat},
					{Label: "Longitude", Value: it.Lon},
					{Label: "Type", Value: it.Type},
				},
			})
		}
	}

	container.Extend(o.Name(), items)
	return nil
}
