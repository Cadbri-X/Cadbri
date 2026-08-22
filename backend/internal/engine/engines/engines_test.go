package engines

import (
	"context"
	"testing"
	"time"

	"cadbri/internal/network"
	"cadbri/internal/results"
	"cadbri/internal/types"
)

func TestGoogleCadbri(t *testing.T) {
	client := network.NewClient(network.ClientOptions{
		Timeout: 5 * time.Second,
	})

	g := NewGoogleEngine()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	container := results.NewResultContainer(nil)
	err := g.Search(ctx, client, types.SearchQuery{Query: "cadbri"}, container)
	res := container.GetOrderedResults()

	t.Logf("Google error: %v", err)
	t.Logf("Google results count: %d", len(res))
	for i, r := range res {
		t.Logf("[%d] %s (%s)", i, r.Title, r.URL)
	}
}
