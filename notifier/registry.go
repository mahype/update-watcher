package notifier

import (
	"fmt"
	"sort"

	"github.com/mahype/update-watcher/config"
)

// FactoryFunc creates a Notifier from a notifier configuration.
type FactoryFunc func(cfg config.NotifierConfig) (Notifier, error)

// NotifierMeta holds display metadata for a notifier type.
type NotifierMeta struct {
	Type        string
	DisplayName string
	Description string
}

var registry = map[string]FactoryFunc{}
var metaRegistry = map[string]NotifierMeta{}

// Register adds a notifier factory to the global registry.
func Register(name string, factory FactoryFunc) {
	registry[name] = factory
}

// RegisterMeta adds display metadata for a notifier type.
func RegisterMeta(meta NotifierMeta) {
	metaRegistry[meta.Type] = meta
}

// Create instantiates a notifier by name using its registered factory.
func Create(name string, cfg config.NotifierConfig) (Notifier, error) {
	factory, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown notifier type: %q", name)
	}
	return factory(cfg)
}

// Available returns all registered notifier names, sorted alphabetically.
func Available() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetMeta returns the metadata for a notifier type.
func GetMeta(notifierType string) (NotifierMeta, bool) {
	meta, ok := metaRegistry[notifierType]
	return meta, ok
}

// AllMeta returns metadata for all registered notifiers, sorted by display name.
func AllMeta() []NotifierMeta {
	metas := make([]NotifierMeta, 0, len(metaRegistry))
	for _, meta := range metaRegistry {
		metas = append(metas, meta)
	}
	sort.Slice(metas, func(i, j int) bool {
		return metas[i].DisplayName < metas[j].DisplayName
	})
	return metas
}
