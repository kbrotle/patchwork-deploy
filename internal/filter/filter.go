package filter

import (
	"fmt"
	"strings"

	"github.com/patchwork-deploy/internal/config"
)

// Criteria holds the filtering parameters for selecting apps.
type Criteria struct {
	Tags   []string
	Hosts  []string
	Apps   []string
}

// Filter selects apps from the config based on the provided criteria.
// If all fields are empty, all app names are returned.
type Filter struct {
	cfg *config.Config
}

// NewFilter creates a new Filter backed by the given config.
func NewFilter(cfg *config.Config) *Filter {
	return &Filter{cfg: cfg}
}

// Apply returns the list of app names that match the criteria.
// An app is included if it satisfies ALL non-empty criteria fields.
func (f *Filter) Apply(c Criteria) ([]string, error) {
	if len(c.Apps) > 0 {
		for _, name := range c.Apps {
			if _, ok := f.cfg.Apps[name]; !ok {
				return nil, fmt.Errorf("filter: unknown app %q", name)
			}
		}
		return c.Apps, nil
	}

	var result []string
	for name, app := range f.cfg.Apps {
		if len(c.Hosts) > 0 && !containsStr(c.Hosts, app.Host) {
			continue
		}
		if len(c.Tags) > 0 && !hasAnyTag(app.Tags, c.Tags) {
			continue
		}
		result = append(result, name)
	}
	return result, nil
}

func containsStr(haystack []string, needle string) bool {
	for _, s := range haystack {
		if strings.EqualFold(s, needle) {
			return true
		}
	}
	return false
}

func hasAnyTag(appTags []string, want []string) bool {
	set := make(map[string]struct{}, len(appTags))
	for _, t := range appTags {
		set[strings.ToLower(t)] = struct{}{}
	}
	for _, w := range want {
		if _, ok := set[strings.ToLower(w)]; ok {
			return true
		}
	}
	return false
}
