package tag_test

import (
	"errors"
	"testing"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/tag"
)

func baseConfig() *config.Config {
	return &config.Config{
		Apps: map[string]config.App{
			"web": {Dir: "/app/web"},
		},
	}
}

func TestNewTagger_NotNil(t *testing.T) {
	tgr := tag.NewTagger(baseConfig(), tag.NewMockStore())
	if tgr == nil {
		t.Fatal("expected non-nil Tagger")
	}
}

func TestTag_UnknownApp(t *testing.T) {
	tgr := tag.NewTagger(baseConfig(), tag.NewMockStore())
	err := tgr.Tag("ghost", "v1", "abc123", nil)
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestTag_EmptyName(t *testing.T) {
	tgr := tag.NewTagger(baseConfig(), tag.NewMockStore())
	err := tgr.Tag("web", "", "abc123", nil)
	if err == nil {
		t.Fatal("expected error for empty tag name")
	}
}

func TestTag_And_Resolve(t *testing.T) {
	tgr := tag.NewTagger(baseConfig(), tag.NewMockStore())
	if err := tgr.Tag("web", "v1.0", "deadbeef", map[string]string{"env": "prod"}); err != nil {
		t.Fatalf("Tag: %v", err)
	}
	entry, err := tgr.Resolve("web", "v1.0")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if entry.Commit != "deadbeef" {
		t.Errorf("expected commit deadbeef, got %q", entry.Commit)
	}
	if entry.Meta["env"] != "prod" {
		t.Errorf("expected meta env=prod, got %q", entry.Meta["env"])
	}
}

func TestResolve_UnknownApp(t *testing.T) {
	tgr := tag.NewTagger(baseConfig(), tag.NewMockStore())
	_, err := tgr.Resolve("ghost", "v1")
	if err == nil {
		t.Fatal("expected error for unknown app")
	}
}

func TestResolve_MissingTag(t *testing.T) {
	tgr := tag.NewTagger(baseConfig(), tag.NewMockStore())
	_, err := tgr.Resolve("web", "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing tag")
	}
}

func TestList_ReturnsTagged(t *testing.T) {
	tgr := tag.NewTagger(baseConfig(), tag.NewMockStore())
	_ = tgr.Tag("web", "v1", "aaa", nil)
	_ = tgr.Tag("web", "v2", "bbb", nil)
	entries, err := tgr.List("web")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestTag_StoreError(t *testing.T) {
	store := tag.NewMockStore()
	store.SaveFn = func(app, tg string, entry tag.Entry) error {
		return errors.New("disk full")
	}
	tgr := tag.NewTagger(baseConfig(), store)
	err := tgr.Tag("web", "v1", "abc", nil)
	if err == nil {
		t.Fatal("expected store error to propagate")
	}
}
