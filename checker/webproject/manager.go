package webproject

import "github.com/mahype/update-watcher/checker"

// PackageManager is the strategy interface for individual package manager handlers.
type PackageManager interface {
	// Name returns the manager identifier (e.g. "npm", "composer").
	Name() string

	// MarkerFiles returns filenames that indicate this manager is in use.
	MarkerFiles() []string

	// CheckOutdated runs the outdated check and returns updates.
	CheckOutdated(project ProjectConfig) ([]checker.Update, error)
}

// SecurityAuditor is an optional interface for managers that support security audits.
type SecurityAuditor interface {
	Audit(project ProjectConfig) ([]checker.Update, error)
}

// Global manager registry (populated by init() in each manager file).
var managerRegistry = map[string]PackageManager{}

// RegisterManager adds a package manager to the registry.
func RegisterManager(mgr PackageManager) {
	managerRegistry[mgr.Name()] = mgr
}

// GetManager returns a manager by name.
func GetManager(name string) (PackageManager, bool) {
	mgr, ok := managerRegistry[name]
	return mgr, ok
}

// AllManagers returns all registered managers.
func AllManagers() []PackageManager {
	managers := make([]PackageManager, 0, len(managerRegistry))
	for _, m := range managerRegistry {
		managers = append(managers, m)
	}
	return managers
}
