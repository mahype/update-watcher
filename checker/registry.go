package checker

import (
	"fmt"
	"sort"

	"github.com/mahype/update-watcher/config"
)

// FactoryFunc creates a Checker from a watcher configuration.
type FactoryFunc func(cfg config.WatcherConfig) (Checker, error)

var registry = map[string]FactoryFunc{}

// Register adds a checker factory to the global registry.
func Register(name string, factory FactoryFunc) {
	registry[name] = factory
}

// Create instantiates a checker by name using its registered factory.
func Create(name string, cfg config.WatcherConfig) (Checker, error) {
	factory, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown checker type: %q", name)
	}
	return factory(cfg)
}

// Available returns all registered checker names, sorted alphabetically.
func Available() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
