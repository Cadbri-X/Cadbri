package search

import (
	"context"
	"strings"
	"sync"
	"time"

	"cadbri/internal/engine"
	"cadbri/internal/network"
	"cadbri/internal/results"
	"cadbri/internal/types"
)

// Orchestrator coordinates concurrent search execution across multiple engines using Goroutines.
type Orchestrator struct {
	client     *network.Client
	registry   *engine.Registry
	webEngines []string
}

// NewOrchestrator creates a new search orchestrator with 26 high-yield web engines.
func NewOrchestrator(client *network.Client, registry *engine.Registry) *Orchestrator {
	if registry == nil {
		registry = engine.GlobalRegistry
	}

	// 26 verified responsive general web engines (each queried once to prevent rate-limiting)
	webEngines := []string{
		"google", "bing", "duckduckgo", "brave", "wikipedia",
		"yahoo", "baidu", "naver", "startpage", "wiby",
		"qwant", "mojeek", "sogou", "so360", "seznam",
		"presearch", "marginalia", "webcrawler", "excite",
		"metacrawler", "ecosia", "swisscows", "dogpile",
		"blackle", "yep", "aol",
	}

	return &Orchestrator{
		client:     client,
		registry:   registry,
		webEngines: webEngines,
	}
}

// Execute parses the search query, selects target engines, runs queries in parallel with ultra-fast sub-second timeouts and aggregates 50-100+ results.
func (o *Orchestrator) Execute(ctx context.Context, query types.SearchQuery) *results.ResultContainer {
	weights := o.registry.GetEngineWeights()
	container := results.NewResultContainer(weights)

	// 1. Check for bang shortcut in query (e.g. "!g facebook", "!gh kubernetes", "!yt interstellar")
	cleanQuery, bangEngine := parseBang(query.Query, o.registry)
	if bangEngine != nil {
		query.Query = cleanQuery
		query.Engines = []string{bangEngine.Name()}
	}

	// 2. Resolve target engines (all 26 engines queried in parallel for massive 50-100+ results)
	targetEngines := o.resolveEngines(query)
	if len(targetEngines) == 0 {
		for _, name := range o.webEngines {
			if e, ok := o.registry.Get(name); ok {
				targetEngines = append(targetEngines, e)
			}
		}
	}

	// 3. Search timeout (1.6 seconds default cap for massive 50+ results aggregation)
	globalTimeout := 1600 * time.Millisecond
	if query.TimeoutLimit > 0 {
		globalTimeout = time.Duration(query.TimeoutLimit * float64(time.Second))
	}

	searchCtx, cancel := context.WithTimeout(ctx, globalTimeout)
	defer cancel()

	var wg sync.WaitGroup
	startTime := time.Now()
	engineDoneCh := make(chan struct{}, len(targetEngines))

	for _, eng := range targetEngines {
		wg.Add(1)
		go func(e engine.Engine) {
			defer wg.Done()
			defer func() {
				select {
				case engineDoneCh <- struct{}{}:
				default:
				}
			}()

			start := time.Now()

			engTimeout := 1400 * time.Millisecond
			if e.Timeout() > 0 && e.Timeout() < engTimeout {
				engTimeout = e.Timeout()
			}

			engCtx, engCancel := context.WithTimeout(searchCtx, engTimeout)
			defer engCancel()

			err := e.Search(engCtx, o.client, query, container)
			elapsed := time.Since(start).Seconds()

			container.AddTiming(types.Timing{
				Engine: e.Name(),
				Total:  elapsed,
			})

			if err != nil {
				if engCtx.Err() == context.DeadlineExceeded || strings.Contains(strings.ToLower(err.Error()), "timeout") {
					container.AddUnresponsiveEngine(e.Name(), "timeout (exceeded execution limit)")
				} else {
					container.AddUnresponsiveEngine(e.Name(), "error: "+err.Error())
				}
			} else if engCtx.Err() == context.DeadlineExceeded {
				container.AddUnresponsiveEngine(e.Name(), "timeout (deadline exceeded)")
			}
		}(eng)
	}

	// Wait for all engines to finish OR early return on sufficient 60+ results
	allDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(allDone)
	}()

	// Soft deadline timer: if we reached 60+ results at 1150ms, don't wait for remaining stragglers
	earlyTimer := time.NewTimer(1150 * time.Millisecond)
	defer earlyTimer.Stop()

	completedEngines := 0
	totalEngines := len(targetEngines)

Loop:
	for {
		select {
		case <-allDone:
			// All engines returned cleanly
			break Loop

		case <-searchCtx.Done():
			// Global timeout reached
			break Loop

		case <-earlyTimer.C:
			// Soft deadline: if we have 60+ results already, return immediately
			if container.ResultsCount() >= 60 {
				break Loop
			}

		case <-engineDoneCh:
			completedEngines++
			elapsed := time.Since(startTime)
			// Target threshold: if we have >= 90 results and at least 500ms elapsed, return immediately
			if container.ResultsCount() >= 90 && elapsed >= 500*time.Millisecond {
				break Loop
			}
			// If all completed
			if completedEngines >= totalEngines {
				break Loop
			}
		}
	}

	// Record engines that produced 0 results
	for _, eng := range targetEngines {
		if !container.HasEngineResults(eng.Name()) && !container.IsEngineUnresponsive(eng.Name()) {
			container.AddUnresponsiveEngine(eng.Name(), "0 results returned or anti-bot challenge")
		}
	}

	// Fallback guarantee: ONLY if all engines yielded 0 results
	if container.ResultsCount() == 0 {
		var fallbackEngines []engine.Engine
		fallbackNames := []string{"brave", "bing", "duckduckgo", "excite"}
		for _, name := range fallbackNames {
			if e, ok := o.registry.Get(name); ok {
				fallbackEngines = append(fallbackEngines, e)
			}
		}

		var fWg sync.WaitGroup
		fallbackCtx, fCancel := context.WithTimeout(ctx, 1000*time.Millisecond)
		defer fCancel()

		for _, fe := range fallbackEngines {
			fWg.Add(1)
			go func(e engine.Engine) {
				defer fWg.Done()
				e.Search(fallbackCtx, o.client, query, container)
			}(fe)
		}
		fWg.Wait()
	}

	return container
}

// resolveEngines returns the list of engine instances that should process this query.
func (o *Orchestrator) resolveEngines(query types.SearchQuery) []engine.Engine {
	seen := make(map[string]struct{})
	var result []engine.Engine

	// 1. If specific engines were requested explicitly
	if len(query.Engines) > 0 {
		for _, name := range query.Engines {
			name = strings.TrimSpace(strings.ToLower(name))
			if name == "" {
				continue
			}
			if e, ok := o.registry.Get(name); ok {
				if _, already := seen[e.Name()]; !already {
					seen[e.Name()] = struct{}{}
					result = append(result, e)
				}
			}
		}
		if len(result) > 0 {
			return result
		}
	}

	// 2. If specific categories were requested (e.g. "videos", "news", "it", "science", "images", "map")
	if len(query.Categories) > 0 {
		for _, cat := range query.Categories {
			cat = strings.TrimSpace(strings.ToLower(cat))
			if cat == "" {
				continue
			}
			for _, e := range o.registry.GetByCategory(cat) {
				if _, already := seen[e.Name()]; !already {
					seen[e.Name()] = struct{}{}
					result = append(result, e)
				}
			}
		}
		if len(result) > 0 {
			return result
		}
	}

	// 3. Default general search: query all 26 verified engines concurrently for rich 50-100+ results
	for _, name := range o.webEngines {
		if e, ok := o.registry.Get(name); ok {
			if _, already := seen[e.Name()]; !already {
				seen[e.Name()] = struct{}{}
				result = append(result, e)
			}
		}
	}

	return result
}

// parseBang extracts "!bang" from the beginning of the query text.
func parseBang(rawQuery string, reg *engine.Registry) (string, engine.Engine) {
	trimmed := strings.TrimSpace(rawQuery)
	if !strings.HasPrefix(trimmed, "!") {
		return rawQuery, nil
	}

	parts := strings.SplitN(trimmed, " ", 2)
	bang := parts[0]
	cleanQuery := ""
	if len(parts) > 1 {
		cleanQuery = strings.TrimSpace(parts[1])
	}

	if e, found := reg.GetByShortcut(bang); found {
		return cleanQuery, e
	}
	if e, found := reg.Get(strings.TrimPrefix(bang, "!")); found {
		return cleanQuery, e
	}

	return rawQuery, nil
}
