package types

// EngineRef references an engine and its associated category.
type EngineRef struct {
	Name     string
	Category string
}

// SearchQuery stores all parsed parameters of a search request.
type SearchQuery struct {
	Query                  string
	Engines                []string
	Categories             []string
	Language               string
	SafeSearch             int // 0: None, 1: Moderate, 2: Strict
	PageNo                 int
	TimeRange              string // "day", "week", "month", "year"
	Format                 string // "json", "csv", "rss"
	TimeoutLimit           float64
	ExternalBang           string
	RedirectToFirstResult  bool
}
