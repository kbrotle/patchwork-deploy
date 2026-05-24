package tag

import (
	"fmt"
	"time"

	"github.com/patchwork-deploy/internal/config"
)

// Store persists and retrieves deployment tags.
type Store interface {
	Save(app, tag string, entry Entry) error
	Load(app, tag string) (Entry, error)
	List(app string) ([]Entry, error)
}

// Entry represents a named deployment tag snapshot.
type Entry struct {
	Tag       string    `json:"tag"`
	App       string    `json:"app"`
	Commit    string    `json:"commit"`
	CreatedAt time.Time `json:"created_at"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// Tagger manages named deployment tags for apps.
type Tagger struct {
	cfg   *config.Config
	store Store
}

// NewTagger creates a new Tagger using the provided config and store.
func NewTagger(cfg *config.Config, store Store) *Tagger {
	return &Tagger{cfg: cfg, store: store}
}

// Tag creates a named tag for the given app at the current time.
func (t *Tagger) Tag(app, tag, commit string, meta map[string]string) error {
	if _, ok := t.cfg.Apps[app]; !ok {
		return fmt.Errorf("tag: unknown app %q", app)
	}
	if tag == "" {
		return fmt.Errorf("tag: tag name must not be empty")
	}
	entry := Entry{
		Tag:       tag,
		App:       app,
		Commit:    commit,
		CreatedAt: time.Now().UTC(),
		Meta:      meta,
	}
	return t.store.Save(app, tag, entry)
}

// Resolve returns the Entry for a given app and tag name.
func (t *Tagger) Resolve(app, tag string) (Entry, error) {
	if _, ok := t.cfg.Apps[app]; !ok {
		return Entry{}, fmt.Errorf("tag: unknown app %q", app)
	}
	return t.store.Load(app, tag)
}

// List returns all tags recorded for an app.
func (t *Tagger) List(app string) ([]Entry, error) {
	if _, ok := t.cfg.Apps[app]; !ok {
		return nil, fmt.Errorf("tag: unknown app %q", app)
	}
	return t.store.List(app)
}
