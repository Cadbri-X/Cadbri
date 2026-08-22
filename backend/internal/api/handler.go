package api

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"cadbri/internal/config"
	"cadbri/internal/engine"
	"cadbri/internal/network"
	"cadbri/internal/search"
	"cadbri/internal/types"
)

// Server sets up the HTTP router and API endpoints.
type Server struct {
	cfg           *config.Config
	orchestrator  *search.Orchestrator
	autocompleter *Autocompleter
	router        chi.Router
}

// NewServer initializes the HTTP API server.
func NewServer(cfg *config.Config, client *network.Client, registry *engine.Registry) *Server {
	s := &Server{
		cfg:           cfg,
		orchestrator:  search.NewOrchestrator(client, registry),
		autocompleter: NewAutocompleter(client),
		router:        chi.NewRouter(),
	}

	s.setupRoutes()
	return s
}

// Router returns the configured chi.Router.
func (s *Server) Router() chi.Router {
	return s.router
}

func (s *Server) setupRoutes() {
	s.router.Use(CORSMiddleware())
	s.router.Use(LoggerMiddleware)

	// Search endpoints
	s.router.Get("/search", s.handleSearch)
	s.router.Post("/search", s.handleSearch)
	s.router.Get("/", s.handleRoot)
	s.router.Post("/", s.handleRoot)

	// Autocompleter endpoints
	s.router.Get("/autocompleter", s.handleAutocompleter)
	s.router.Post("/autocompleter", s.handleAutocompleter)

	// Health & Info endpoints
	s.router.Get("/healthz", s.handleHealth)
	s.router.Get("/config", s.handleConfig)
	s.router.Get("/stats", s.handleStats)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q != "" {
		s.handleSearch(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":      s.cfg.General.InstanceName,
		"version":   "1.0.0-go",
		"status":    "running",
		"endpoints": []string{"/search?q=...", "/autocompleter?q=...", "/healthz", "/stats"},
	})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	_ = r.ParseForm()
	qText := strings.TrimSpace(r.FormValue("q"))
	if qText == "" {
		http.Error(w, `{"error": "missing query parameter 'q'"}`, http.StatusBadRequest)
		return
	}

	format := strings.ToLower(strings.TrimSpace(r.FormValue("format")))
	if format == "" {
		format = "json"
	}

	// Parse engines (comma-separated or multiple params)
	var enginesList []string
	if rawEngines := r.FormValue("engines"); rawEngines != "" {
		for _, e := range strings.Split(rawEngines, ",") {
			e = strings.TrimSpace(e)
			if e != "" {
				enginesList = append(enginesList, e)
			}
		}
	}
	if len(r.Form["engines"]) > 1 {
		for _, e := range r.Form["engines"] {
			e = strings.TrimSpace(e)
			if e != "" {
				enginesList = append(enginesList, e)
			}
		}
	}

	// Parse categories
	var categoriesList []string
	if rawCats := r.FormValue("categories"); rawCats != "" {
		for _, c := range strings.Split(rawCats, ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				categoriesList = append(categoriesList, c)
			}
		}
	}

	// Parse pagination
	pageNo := 1
	if pStr := r.FormValue("pageno"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			pageNo = p
		}
	}

	// Parse SafeSearch
	safeSearch := s.cfg.Search.SafeSearch
	if ssStr := r.FormValue("safesearch"); ssStr != "" {
		if ss, err := strconv.Atoi(ssStr); err == nil && ss >= 0 && ss <= 2 {
			safeSearch = ss
		}
	}

	// Parse timeout limit
	timeoutLimit := s.cfg.Outgoing.RequestTimeout
	if tStr := r.FormValue("timeout_limit"); tStr != "" {
		if t, err := strconv.ParseFloat(tStr, 64); err == nil && t > 0 {
			timeoutLimit = t
		}
	}

	searchQuery := types.SearchQuery{
		Query:        qText,
		Engines:      enginesList,
		Categories:   categoriesList,
		Language:     r.FormValue("language"),
		SafeSearch:   safeSearch,
		PageNo:       pageNo,
		TimeRange:    r.FormValue("time_range"),
		Format:       format,
		TimeoutLimit: timeoutLimit,
	}

	// Execute search concurrently
	container := s.orchestrator.Execute(r.Context(), searchQuery)
	response := container.BuildResponse(qText)

	// Format response
	switch format {
	case "csv":
		s.renderCSV(w, response)
	case "rss":
		s.renderRSS(w, response)
	default:
		// Default to JSON
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(response)
	}
}

func (s *Server) renderCSV(w http.ResponseWriter, resp types.SearchResponse) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"cadbri_%s.csv\"", resp.Query))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	_ = writer.Write([]string{"title", "url", "content", "engine", "score"})
	for _, res := range resp.Results {
		_ = writer.Write([]string{
			res.Title,
			res.URL,
			res.Content,
			res.Engine,
			fmt.Sprintf("%f", res.Score),
		})
	}
}

type rssFeed struct {
	XMLName xml.Name  `xml:"rss"`
	Version string    `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

func (s *Server) renderRSS(w http.ResponseWriter, resp types.SearchResponse) {
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")

	feed := rssFeed{
		Version: "2.0",
		Channel: rssChannel{
			Title:       fmt.Sprintf("Search results for %s", resp.Query),
			Link:        "http://localhost:2222",
			Description: fmt.Sprintf("%d results for query '%s'", resp.NumberOfResults, resp.Query),
		},
	}

	for _, res := range resp.Results {
		feed.Channel.Items = append(feed.Channel.Items, rssItem{
			Title:       res.Title,
			Link:        res.URL,
			Description: res.Content,
		})
	}

	w.Write([]byte(xml.Header))
	xml.NewEncoder(w).Encode(feed)
}

func (s *Server) handleAutocompleter(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	qText := strings.TrimSpace(r.FormValue("q"))
	backend := r.FormValue("backend")
	if backend == "" {
		backend = s.cfg.Search.Autocomplete
	}

	suggestions := s.autocompleter.Complete(r.Context(), qText, backend)
	if suggestions == nil {
		suggestions = []string{}
	}

	// Check if OpenSearch format was requested or standard XHR
	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(suggestions)
	} else {
		// OpenSearch / Cadbri default: [query, [suggestions...]]
		w.Header().Set("Content-Type", "application/x-suggestions+json; charset=utf-8")
		json.NewEncoder(w).Encode([]interface{}{qText, suggestions})
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "ok",
		"engines_total": len(engine.GlobalRegistry.GetAll()),
		"timestamp":     r.Context().Value("time"),
	})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.cfg)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	engines := engine.GlobalRegistry.GetAll()
	engineStats := make([]map[string]interface{}, 0, len(engines))

	for _, e := range engines {
		engineStats = append(engineStats, map[string]interface{}{
			"name":       e.Name(),
			"categories": e.Categories(),
			"weight":     e.Weight(),
			"timeout_s":  e.Timeout().Seconds(),
			"paging":     e.SupportsPaging(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"engines": engineStats,
		"total":   len(engines),
	})
}
