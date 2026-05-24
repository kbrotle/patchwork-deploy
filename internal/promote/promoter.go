package promote

import (
	"fmt"

	"github.com/yourorg/patchwork-deploy/internal/config"
	"github.com/yourorg/patchwork-deploy/internal/tag"
	"github.com/yourorg/patchwork-deploy/internal/version"
)

// Promoter copies a tagged release from one app environment to another,
// recording the promoted version in the version tracker.
type Promoter struct {
	cfg     *config.Config
	tagger  *tag.Tagger
	tracker *version.Tracker
}

// NewPromoter returns a Promoter wired up with the given dependencies.
func NewPromoter(cfg *config.Config, tagger *tag.Tagger, tracker *version.Tracker) *Promoter {
	return &Promoter{cfg: cfg, tagger: tagger, tracker: tracker}
}

// Promote resolves the named tag on srcApp and records it as the current
// version on dstApp. Both apps must exist in the config.
func (p *Promoter) Promote(srcApp, dstApp, tagName string) error {
	if err := p.requireApp(srcApp); err != nil {
		return err
	}
	if err := p.requireApp(dstApp); err != nil {
		return err
	}
	if tagName == "" {
		return fmt.Errorf("promote: tag name must not be empty")
	}

	resolved, err := p.tagger.Resolve(srcApp, tagName)
	if err != nil {
		return fmt.Errorf("promote: resolve tag %q on %q: %w", tagName, srcApp, err)
	}

	if err := p.tracker.Record(dstApp, resolved); err != nil {
		return fmt.Errorf("promote: record version on %q: %w", dstApp, err)
	}

	return nil
}

func (p *Promoter) requireApp(name string) error {
	for _, a := range p.cfg.Apps {
		if a.Name == name {
			return nil
		}
	}
	return fmt.Errorf("promote: unknown app %q", name)
}
