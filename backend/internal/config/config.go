package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FlexibleStringSlice handles YAML fields that can be either a string or a []string.
type FlexibleStringSlice []string

func (f *FlexibleStringSlice) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		*f = []string{value.Value}
		return nil
	case yaml.SequenceNode:
		var items []string
		if err := value.Decode(&items); err != nil {
			// If items in the sequence are not all strings, try interface decode
			var raw []interface{}
			if err2 := value.Decode(&raw); err2 != nil {
				return err
			}
			for _, v := range raw {
				if s, ok := v.(string); ok {
					items = append(items, s)
				}
			}
		}
		*f = items
		return nil
	default:
		*f = nil
		return nil
	}
}

// FlexibleString handles YAML fields that can be a string, a list, or a number.
type FlexibleString string

func (f *FlexibleString) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		*f = FlexibleString(value.Value)
		return nil
	case yaml.SequenceNode:
		// Some fields like base_url can be a list; take the first element
		var items []string
		if err := value.Decode(&items); err == nil && len(items) > 0 {
			*f = FlexibleString(items[0])
		}
		return nil
	default:
		*f = ""
		return nil
	}
}

type Config struct {
	General  GeneralSettings  `yaml:"general"`
	Server   ServerSettings   `yaml:"server"`
	Search   SearchSettings   `yaml:"search"`
	Outgoing OutgoingSettings `yaml:"outgoing"`
	Engines  []EngineConfig   `yaml:"engines"`
}

type GeneralSettings struct {
	InstanceName  string `yaml:"instance_name"`
	Debug         bool   `yaml:"debug"`
	EnableMetrics bool   `yaml:"enable_metrics"`
}

type ServerSettings struct {
	Port        int    `yaml:"port"`
	BindAddress string `yaml:"bind_address"`
	SecretKey   string `yaml:"secret_key"`
	BaseURL     string `yaml:"base_url"`
}

type SearchSettings struct {
	SafeSearch   int      `yaml:"safe_search"`
	Autocomplete string   `yaml:"autocomplete"`
	DefaultLang  string   `yaml:"default_lang"`
	MaxPage      int      `yaml:"max_page"`
	Formats      []string `yaml:"formats"`
}

type OutgoingSettings struct {
	RequestTimeout    float64           `yaml:"request_timeout"`
	MaxRequestTimeout float64           `yaml:"max_request_timeout"`
	PoolConnections   int               `yaml:"pool_connections"`
	PoolMaxSize       int               `yaml:"pool_maxsize"`
	EnableHTTP2       bool              `yaml:"enable_http2"`
	UserAgent         string            `yaml:"useragent"`
	Proxies           map[string]string `yaml:"proxies"`
	UsingTorProxy     bool              `yaml:"using_tor_proxy"`
	Verify            bool              `yaml:"verify"`
}

type EngineConfig struct {
	Name       string              `yaml:"name"`
	Engine     FlexibleString      `yaml:"engine"`
	Shortcut   FlexibleString      `yaml:"shortcut"`
	Categories FlexibleStringSlice `yaml:"categories"`
	Timeout    float64             `yaml:"timeout"`
	Weight     float64             `yaml:"weight"`
	Disabled   bool                `yaml:"disabled"`
	Inactive   bool                `yaml:"inactive"`
	BaseURL    FlexibleString      `yaml:"base_url"`
	APIKey     FlexibleString      `yaml:"api_key"`
}

// DefaultConfig returns default configuration parameters.
func DefaultConfig() *Config {
	return &Config{
		General: GeneralSettings{
			InstanceName:  "Cadbri Search",
			Debug:         false,
			EnableMetrics: false,
		},
		Server: ServerSettings{
			Port:        2222,
			BindAddress: "0.0.0.0",
			SecretKey:   "secret_key_cadbri",
			BaseURL:     "",
		},
		Search: SearchSettings{
			SafeSearch:   0,
			Autocomplete: "duckduckgo",
			DefaultLang:  "all",
			MaxPage:      0,
			Formats:      []string{"html", "csv", "json", "rss"},
		},
		Outgoing: OutgoingSettings{
			RequestTimeout:    3.0,
			MaxRequestTimeout: 6.0,
			PoolConnections:   100,
			PoolMaxSize:       100,
			EnableHTTP2:       true,
			UserAgent:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
			Verify:            true,
		},
		Engines: []EngineConfig{},
	}
}

// LoadConfig loads configuration from the given file path or common fallback paths.
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	candidatePaths := []string{
		path,
		"settings.yml",
		"backend/settings.yml",
		"../backend/settings.yml",
		"/etc/cadbri/settings.yml",
	}

	var foundPath string
	for _, p := range candidatePaths {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			foundPath = p
			break
		}
	}

	if foundPath == "" {
		fmt.Println("[Config] No settings.yml found, using default configuration.")
		return cfg, nil
	}

	absPath, _ := filepath.Abs(foundPath)
	fmt.Printf("[Config] Loading settings from: %s\n", absPath)

	data, err := os.ReadFile(foundPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file %s: %w", foundPath, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML settings %s: %w", foundPath, err)
	}

	fmt.Printf("[Config] Loaded %d engine configs from settings.yml\n", len(cfg.Engines))

	return cfg, nil
}
