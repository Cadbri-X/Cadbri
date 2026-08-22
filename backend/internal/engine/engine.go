package engine

import (
	"context"
	"time"

	"cadbri/internal/network"
	"cadbri/internal/results"
	"cadbri/internal/types"
)

// Engine defines the contract for all upstream search engine scrapers and APIs.
type Engine interface {
	Name() string
	Categories() []string
	Weight() float64
	Timeout() time.Duration
	SupportsPaging() bool
	Search(ctx context.Context, client *network.Client, query types.SearchQuery, container *results.ResultContainer) error
}

// BaseEngine provides standard default properties for search engines.
type BaseEngine struct {
	EngineName     string
	EngineCategories []string
	EngineWeight   float64
	EngineTimeout  time.Duration
	CanPage        bool
}

func (b *BaseEngine) Name() string {
	return b.EngineName
}

func (b *BaseEngine) Categories() []string {
	return b.EngineCategories
}

func (b *BaseEngine) Weight() float64 {
	if b.EngineWeight <= 0 {
		return 1.0
	}
	return b.EngineWeight
}

func (b *BaseEngine) Timeout() time.Duration {
	if b.EngineTimeout <= 0 {
		return 3 * time.Second
	}
	return b.EngineTimeout
}

func (b *BaseEngine) SupportsPaging() bool {
	return b.CanPage
}
