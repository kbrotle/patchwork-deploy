package promote_test

import (
	"testing"

	"github.com/yourorg/patchwork-deploy/internal/config"
	"github.com/yourorg/patchwork-deploy/internal/promote"
	"github.com/yourorg/patchwork-deploy/internal/tag"
	"github.com/yourorg/patchwork-deploy/internal/version"
)

func baseConfig() *config.Config {
	return &config.Config{
		Apps: []config.App{
			{Name: "api"},
			{Name: "worker"},
		},
	}
}

func makePromoter(t *testing.T) (*promote.Promoter, *promote.MockTagStore, *promote.MockVersionStore) {
	t.Helper()
	cfg := baseConfig()
	tagStore := promote.NewMockTagStore()
	versionStore := promote.NewMockVersionStore()
	tagger := tag.NewTagger(cfg, tagStore)
	tracker := version.NewTracker(cfg, versionStore)
	return promote.NewPromoter(cfg, tagger, tracker), tagStore, versionStore
}

func TestNewPromoter_NotNil(t *testing.T) {
	p, _, _ := makePromoter(t)
	if p == nil {
		t.Fatal("expected non-nil Promoter")
	}
}

func TestPromote_UnknownSrcApp(t *testing.T) {
	p, _, _ := makePromoter(t)
	if err := p.Promote("ghost", "worker", "v1"); err == nil {
		t.Fatal("expected error for unknown src app")
	}
}

func TestPromote_UnknownDstApp(t *testing.T) {
	p, _, _ := makePromoter(t)
	if err := p.Promote("api", "ghost", "v1"); err == nil {
		t.Fatal("expected error for unknown dst app")
	}
}

func TestPromote_EmptyTag(t *testing.T) {
	p, _, _ := makePromoter(t)
	if err := p.Promote("api", "worker", ""); err == nil {
		t.Fatal("expected error for empty tag")
	}
}

func TestPromote_TagNotFound(t *testing.T) {
	p, _, _ := makePromoter(t)
	if err := p.Promote("api", "worker", "missing-tag"); err == nil {
		t.Fatal("expected error when tag does not exist on src app")
	}
}

func TestPromote_Success(t *testing.T) {
	p, tagStore, versionStore := makePromoter(t)

	if err := tagStore.Save("api", "stable", "sha256:abc123"); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := p.Promote("api", "worker", "stable"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := versionStore.Load("worker")
	if err != nil {
		t.Fatalf("version not recorded: %v", err)
	}
	if got != "sha256:abc123" {
		t.Errorf("version = %q, want %q", got, "sha256:abc123")
	}
}
