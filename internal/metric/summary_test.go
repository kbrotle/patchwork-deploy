package metric_test

import (
	"testing"

	"github.com/patchwork-deploy/internal/metric"
)

func makeEntries(app, name string, vals ...float64) []metric.Entry {
	var out []metric.Entry
	for _, v := range vals {
		out = append(out, metric.Entry{App: app, Name: name, Value: v})
	}
	return out
}

func TestSummarise_NoEntries(t *testing.T) {
	_, err := metric.Summarise("web", "cpu", nil)
	if err == nil {
		t.Fatal("expected error for empty entries")
	}
}

func TestSummarise_NoMatchingName(t *testing.T) {
	entries := makeEntries("web", "mem", 100, 200)
	_, err := metric.Summarise("web", "cpu", entries)
	if err == nil {
		t.Fatal("expected error when name not found")
	}
}

func TestSummarise_CorrectStats(t *testing.T) {
	entries := makeEntries("web", "cpu", 10, 20, 30)
	s, err := metric.Summarise("web", "cpu", entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Count != 3 {
		t.Errorf("expected count=3, got %d", s.Count)
	}
	if s.Sum != 60 {
		t.Errorf("expected sum=60, got %f", s.Sum)
	}
	if s.Min != 10 {
		t.Errorf("expected min=10, got %f", s.Min)
	}
	if s.Max != 30 {
		t.Errorf("expected max=30, got %f", s.Max)
	}
	if s.Avg != 20 {
		t.Errorf("expected avg=20, got %f", s.Avg)
	}
}

func TestSummarise_FiltersOtherNames(t *testing.T) {
	entries := append(
		makeEntries("web", "cpu", 5, 15),
		makeEntries("web", "mem", 999)...,
	)
	s, err := metric.Summarise("web", "cpu", entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Count != 2 {
		t.Errorf("expected count=2, got %d", s.Count)
	}
	if s.Max != 15 {
		t.Errorf("expected max=15, got %f", s.Max)
	}
}
