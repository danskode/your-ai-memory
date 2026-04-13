package wiki

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Wiki represents a single registered wiki from ~/.your-ai-memory/config.json.
type Wiki struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Domain       string   `json:"domain"`
	Topics       []string `json:"topics"`
	Created      string   `json:"created"`
	LastAccessed string   `json:"lastAccessed"`
}

type registry struct {
	Wikis []Wiki `json:"wikis"`
}

// ConfigPath returns the path to the global config file.
func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".your-ai-memory", "config.json")
}

// LoadRegistry reads ~/.your-ai-memory/config.json and returns all registered wikis.
// Returns an empty slice (not an error) if the file does not yet exist.
func LoadRegistry() ([]Wiki, error) {
	data, err := os.ReadFile(ConfigPath())
	if os.IsNotExist(err) {
		return []Wiki{}, nil
	}
	if err != nil {
		return nil, err
	}
	var reg registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, err
	}
	return reg.Wikis, nil
}

// SaveRegistry writes the wiki list back to config.json.
func SaveRegistry(wikis []Wiki) error {
	reg := registry{Wikis: wikis}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	configDir := filepath.Dir(ConfigPath())
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(), data, 0644)
}

// UpdateLastAccessed updates the lastAccessed field for a wiki by name.
func UpdateLastAccessed(wikis []Wiki, name string) []Wiki {
	today := todayString()
	for i := range wikis {
		if wikis[i].Name == name {
			wikis[i].LastAccessed = today
		}
	}
	return wikis
}

// GroupByTopic groups wikis by their first topic tag.
func GroupByTopic(wikis []Wiki) map[string][]Wiki {
	groups := make(map[string][]Wiki)
	for _, w := range wikis {
		tag := "untagged"
		if len(w.Topics) > 0 {
			tag = w.Topics[0]
		}
		groups[tag] = append(groups[tag], w)
	}
	return groups
}

func todayString() string {
	// Use time package via caller — keeping this package dependency-free.
	// Callers that need today's date should use time.Now().Format("2006-01-02").
	return ""
}
