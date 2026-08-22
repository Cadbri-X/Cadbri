package types

// SearchResult represents an individual search result item in clean modern JSON format.
type SearchResult struct {
	URL           string   `json:"url"`
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	ParsedURL     []string `json:"parsed_url"`
	Engines       []string `json:"engines"`
	Category      string   `json:"category"`
	Thumbnail     string   `json:"thumbnail,omitempty"`
	ImgSrc        string   `json:"img_src,omitempty"`
	PublishedDate *string  `json:"publishedDate,omitempty"`
	PubDate       string   `json:"pubdate,omitempty"`
	Author        string   `json:"author,omitempty"`
	Views         string   `json:"views,omitempty"`
	Length        *int     `json:"length,omitempty"`
	IframeSrc     string   `json:"iframe_src,omitempty"`
	AudioSrc      string   `json:"audio_src,omitempty"`
	Metadata      string   `json:"metadata,omitempty"`

	// Internal processing fields (hidden from clean JSON output)
	Engine     string  `json:"-"`
	Template   string  `json:"-"`
	Priority   string  `json:"-"`
	Positions  []int   `json:"-"`
	Score      float64 `json:"-"`
	OpenGroup  bool    `json:"-"`
	CloseGroup bool    `json:"-"`
}

// Answer represents instant answer data (e.g. calculator, definitions, Wikipedia quick answer).
type Answer struct {
	URL      string `json:"url,omitempty"`
	Answer   string `json:"answer"`
	Engine   string `json:"engine,omitempty"`
	Template string `json:"template,omitempty"`
}

// Infobox represents sidebar entity information.
type Infobox struct {
	Infobox    string             `json:"infobox"`
	ID         string             `json:"id,omitempty"`
	Content    string             `json:"content,omitempty"`
	ImgSrc     string             `json:"img_src,omitempty"`
	Engine     string             `json:"engine,omitempty"`
	URLs       []InfoboxURL       `json:"urls,omitempty"`
	Attributes []InfoboxAttribute `json:"attributes,omitempty"`
}

type InfoboxURL struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type InfoboxAttribute struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// Timing records execution timing for a specific engine.
type Timing struct {
	Engine string  `json:"engine"`
	Total  float64 `json:"total"`
	Load   float64 `json:"load,omitempty"`
}

// UnresponsiveEngine records an engine that failed or timed out.
type UnresponsiveEngine struct {
	Engine    string `json:"engine"`
	ErrorType string `json:"error_type"`
}

// SearchResponse represents the complete JSON response returned by `/api/search?format=json`.
type SearchResponse struct {
	Query               string         `json:"query"`
	NumberOfResults     int64          `json:"number_of_results"`
	Results             []SearchResult `json:"results"`
	Answers             []Answer       `json:"answers"`
	Corrections         []string       `json:"corrections"`
	Infoboxes           []Infobox      `json:"infoboxes"`
	Suggestions         []string       `json:"suggestions"`
	UnresponsiveEngines [][]string     `json:"unresponsive_engines"`
}
