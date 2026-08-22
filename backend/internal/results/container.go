package results

import (
	"net/url"
	"sort"
	"strings"
	"sync"

	"cadbri/internal/types"
)

// ResultContainer collects, deduplicates, scores, and sorts search results from multiple engines.
type ResultContainer struct {
	mu                  sync.RWMutex
	mainResultsMap      map[string]*types.SearchResult
	mainResultsOrder    []string
	infoboxes           []types.Infobox
	answers             []types.Answer
	suggestions         map[string]struct{}
	corrections         map[string]struct{}
	unresponsiveEngines []types.UnresponsiveEngine
	timings             []types.Timing
	numberOfResults     int64
	engineWeights       map[string]float64
}

// NewResultContainer creates an empty ResultContainer.
func NewResultContainer(engineWeights map[string]float64) *ResultContainer {
	if engineWeights == nil {
		engineWeights = make(map[string]float64)
	}
	return &ResultContainer{
		mainResultsMap:      make(map[string]*types.SearchResult),
		mainResultsOrder:    make([]string, 0),
		infoboxes:           make([]types.Infobox, 0),
		answers:             make([]types.Answer, 0),
		suggestions:         make(map[string]struct{}),
		corrections:         make(map[string]struct{}),
		unresponsiveEngines: make([]types.UnresponsiveEngine, 0),
		timings:             make([]types.Timing, 0),
		engineWeights:       engineWeights,
	}
}

// NormalizeURL strips tracking queries, fragments, and standardizes trailing slashes.
func NormalizeURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	u.Fragment = "" // remove anchor #

	// Strip common tracking query params
	if u.RawQuery != "" {
		q := u.Query()
		trackingKeys := []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content", "fbclid", "gclid", "msclkid"}
		for _, key := range trackingKeys {
			q.Del(key)
		}
		u.RawQuery = q.Encode()
	}

	u.Host = strings.ToLower(u.Host)
	u.Path = strings.TrimRight(u.Path, "/")
	if u.Path == "" {
		u.Path = "/"
	}

	return u.String()
}

// ParseURLParts splits a URL into [scheme, host, path]
func ParseURLParts(rawURL string) []string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return []string{"https", "", "/"}
	}
	path := u.Path
	if path == "" {
		path = "/"
	}
	return []string{u.Scheme, u.Host, path}
}

// Extend merges a batch of results from a specific engine into the container.
func (rc *ResultContainer) Extend(engineName string, items []types.SearchResult) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	for i, item := range items {
		if item.URL == "" && item.Title == "" {
			continue
		}

		normURL := NormalizeURL(item.URL)
		if normURL == "" {
			continue
		}

		position := i + 1
		if len(item.Positions) > 0 {
			position = item.Positions[0]
		}

		existing, found := rc.mainResultsMap[normURL]
		if found {
			// Merge engine reference
			hasEngine := false
			for _, e := range existing.Engines {
				if e == engineName {
					hasEngine = true
					break
				}
			}
			if !hasEngine {
				existing.Engines = append(existing.Engines, engineName)
				existing.Positions = append(existing.Positions, position)
			}

			// Keep longer content description if new one is more informative
			if len(item.Content) > len(existing.Content) {
				existing.Content = item.Content
			}
			if existing.Thumbnail == "" && item.Thumbnail != "" {
				existing.Thumbnail = item.Thumbnail
			}
			if existing.ImgSrc == "" && item.ImgSrc != "" {
				existing.ImgSrc = item.ImgSrc
			}
			if existing.PublishedDate == nil && item.PublishedDate != nil {
				existing.PublishedDate = item.PublishedDate
			}
			if existing.PubDate == "" && item.PubDate != "" {
				existing.PubDate = item.PubDate
			}

			// Recalculate score with boosted weight for multi-engine appearance
			rc.calculateScore(existing)
		} else {
			newItem := item
			newItem.URL = normURL
			newItem.Engine = engineName
			newItem.Engines = []string{engineName}
			newItem.Positions = []int{position}
			if newItem.Template == "" {
				newItem.Template = "default.html"
			}
			if len(newItem.ParsedURL) == 0 {
				newItem.ParsedURL = ParseURLParts(newItem.URL)
			}

			rc.calculateScore(&newItem)
			rc.mainResultsMap[normURL] = &newItem
			rc.mainResultsOrder = append(rc.mainResultsOrder, normURL)
		}
	}
}

// calculateScore computes the ranking score for a result.
func (rc *ResultContainer) calculateScore(r *types.SearchResult) {
	weight := 1.0
	for _, eng := range r.Engines {
		if w, ok := rc.engineWeights[eng]; ok && w > 0 {
			weight *= w
		}
	}

	weight *= float64(len(r.Positions))
	score := 0.0

	for _, pos := range r.Positions {
		if r.Priority == "low" {
			continue
		}
		if r.Priority == "high" {
			score += weight
		} else {
			if pos > 0 {
				score += weight / float64(pos)
			} else {
				score += weight
			}
		}
	}

	r.Score = score
}

// AddAnswer adds an instant answer.
func (rc *ResultContainer) AddAnswer(ans types.Answer) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.answers = append(rc.answers, ans)
}

// AddInfobox adds and deduplicates an infobox by entity title.
func (rc *ResultContainer) AddInfobox(info types.Infobox) {
	info.Infobox = strings.TrimSpace(info.Infobox)
	if info.Infobox == "" {
		return
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()

	normTitle := strings.ToLower(info.Infobox)
	for i, existing := range rc.infoboxes {
		if strings.ToLower(strings.TrimSpace(existing.Infobox)) == normTitle {
			// Merge into existing: prefer richer content, valid image, or URLs
			if len(info.Content) > len(existing.Content) {
				rc.infoboxes[i].Content = info.Content
			}
			if existing.ImgSrc == "" && info.ImgSrc != "" {
				rc.infoboxes[i].ImgSrc = info.ImgSrc
			}
			if len(existing.URLs) == 0 && len(info.URLs) > 0 {
				rc.infoboxes[i].URLs = info.URLs
			}
			if len(existing.Attributes) == 0 && len(info.Attributes) > 0 {
				rc.infoboxes[i].Attributes = info.Attributes
			}
			return
		}
	}

	rc.infoboxes = append(rc.infoboxes, info)
}

// AddSuggestion adds a search suggestion.
func (rc *ResultContainer) AddSuggestion(suggestion string) {
	suggestion = strings.TrimSpace(suggestion)
	if suggestion == "" {
		return
	}
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.suggestions[suggestion] = struct{}{}
}

// AddCorrection adds a spelling correction.
func (rc *ResultContainer) AddCorrection(correction string) {
	correction = strings.TrimSpace(correction)
	if correction == "" {
		return
	}
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.corrections[correction] = struct{}{}
}

// SetNumberOfResults updates the estimated total result count.
func (rc *ResultContainer) SetNumberOfResults(n int64) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if n > rc.numberOfResults {
		rc.numberOfResults = n
	}
}

// AddTiming records engine timing.
func (rc *ResultContainer) AddTiming(t types.Timing) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.timings = append(rc.timings, t)
}

// AddUnresponsiveEngine records an engine timeout or failure.
func (rc *ResultContainer) AddUnresponsiveEngine(name, errType string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	for _, u := range rc.unresponsiveEngines {
		if u.Engine == name {
			return // already recorded
		}
	}

	rc.unresponsiveEngines = append(rc.unresponsiveEngines, types.UnresponsiveEngine{
		Engine:    name,
		ErrorType: errType,
	})
}

// IsEngineUnresponsive checks if an engine is already in the unresponsive list.
func (rc *ResultContainer) IsEngineUnresponsive(name string) bool {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	for _, u := range rc.unresponsiveEngines {
		if u.Engine == name {
			return true
		}
	}
	return false
}

// HasEngineResults checks if any result was contributed by the given engine.
func (rc *ResultContainer) HasEngineResults(name string) bool {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	for _, res := range rc.mainResultsMap {
		for _, e := range res.Engines {
			if e == name {
				return true
			}
		}
	}
	return false
}

// ResultsCount returns the current number of unique results collected so far.
func (rc *ResultContainer) ResultsCount() int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return len(rc.mainResultsMap)
}

// GetOrderedResults returns deduplicated search results sorted by score in descending order.
func (rc *ResultContainer) GetOrderedResults() []types.SearchResult {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	results := make([]types.SearchResult, 0, len(rc.mainResultsMap))
	for _, res := range rc.mainResultsMap {
		if len(res.ParsedURL) == 0 {
			res.ParsedURL = ParseURLParts(res.URL)
		}
		results = append(results, *res)
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// BuildResponse compiles the final SearchResponse payload.
func (rc *ResultContainer) BuildResponse(query string) types.SearchResponse {
	results := rc.GetOrderedResults()

	rc.mu.RLock()
	defer rc.mu.RUnlock()

	answers := rc.answers
	if answers == nil {
		answers = make([]types.Answer, 0)
	}

	infoboxes := rc.infoboxes
	if infoboxes == nil {
		infoboxes = make([]types.Infobox, 0)
	}

	suggestions := make([]string, 0, len(rc.suggestions))
	for s := range rc.suggestions {
		suggestions = append(suggestions, s)
	}

	corrections := make([]string, 0, len(rc.corrections))
	for c := range rc.corrections {
		corrections = append(corrections, c)
	}

	unresponsive := make([][]string, 0, len(rc.unresponsiveEngines))
	for _, u := range rc.unresponsiveEngines {
		unresponsive = append(unresponsive, []string{u.Engine, u.ErrorType})
	}

	totalResults := rc.numberOfResults
	if totalResults == 0 {
		totalResults = int64(len(results))
	}

	return types.SearchResponse{
		Query:               query,
		NumberOfResults:     totalResults,
		Results:             results,
		Answers:             answers,
		Corrections:         corrections,
		Infoboxes:           infoboxes,
		Suggestions:         suggestions,
		UnresponsiveEngines: unresponsive,
	}
}
