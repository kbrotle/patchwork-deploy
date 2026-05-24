package metric

import (
	"fmt"
	"math"
)

// Summary holds aggregate statistics for a set of metric entries.
type Summary struct {
	App   string
	Name  string
	Count int
	Sum   float64
	Min   float64
	Max   float64
	Avg   float64
}

// Summarise computes aggregate stats for a named metric across the given entries.
// Returns an error if no matching entries are found.
func Summarise(app, name string, entries []Entry) (Summary, error) {
	var matched []Entry
	for _, e := range entries {
		if e.Name == name {
			matched = append(matched, e)
		}
	}
	if len(matched) == 0 {
		return Summary{}, fmt.Errorf("metric: no entries for app=%q name=%q", app, name)
	}
	s := Summary{
		App:  app,
		Name: name,
		Min:  math.MaxFloat64,
		Max:  -math.MaxFloat64,
	}
	for _, e := range matched {
		s.Count++
		s.Sum += e.Value
		if e.Value < s.Min {
			s.Min = e.Value
		}
		if e.Value > s.Max {
			s.Max = e.Value
		}
	}
	if s.Count > 0 {
		s.Avg = s.Sum / float64(s.Count)
	}
	return s, nil
}
