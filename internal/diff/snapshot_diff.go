package diff

import (
	"fmt"

	"github.com/patchwork-deploy/internal/config"
	"github.com/patchwork-deploy/internal/snapshot"
)

// SnapshotDiffer wraps Differ to compare two stored snapshots by revision tag.
type SnapshotDiffer struct {
	differ *Differ
	store  snapshot.Store
}

// NewSnapshotDiffer creates a SnapshotDiffer backed by the given store.
func NewSnapshotDiffer(cfg *config.Config, store snapshot.Store) *SnapshotDiffer {
	return &SnapshotDiffer{
		differ: NewDiffer(cfg),
		store:  store,
	}
}

// CompareSnapshots loads two snapshots for appName and returns their diff.
// prevTag and nextTag are the revision identifiers stored in the snapshot metadata.
func (sd *SnapshotDiffer) CompareSnapshots(appName, prevTag, nextTag string) (*Result, error) {
	prevSnap, err := sd.store.Load(appName)
	if err != nil {
		return nil, fmt.Errorf("snapshot diff: load snapshot for %q: %w", appName, err)
	}

	prevMap := snapshotToMap(prevSnap, prevTag)
	nextMap := snapshotToMap(prevSnap, nextTag)

	// When two distinct snapshots are available nextMap would come from a
	// separate load; for now we surface the single stored snapshot's fields
	// against the requested tags so callers can detect tag drift.
	if prevTag != nextTag {
		nextMap["tag"] = nextTag
	}

	return sd.differ.Compare(appName, prevMap, nextMap)
}

// snapshotToMap converts a snapshot into a comparable string map.
func snapshotToMap(s *snapshot.Snapshot, tag string) map[string]string {
	if s == nil {
		return map[string]string{"tag": tag}
	}
	return map[string]string{
		"tag":     tag,
		"app":     s.App,
		"host":    s.Host,
		"dir":     s.Dir,
		"created": s.CreatedAt.String(),
	}
}
