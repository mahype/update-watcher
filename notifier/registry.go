package notifier

import (
	"fmt"
	"sort"

	"github.com/mahype/update-watcher/config"
)

// FactoryFunc creates a Notifier from a notifier configuration.
type FactoryFunc func(cfg config.NotifierConfig) (Notifier, error)

var registry = map[string]FactoryFunc{}

// Register adds a notifier factory to the global registry.
func Register(name string, factory FactoryFunc) {
	registry[name] = factory
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
